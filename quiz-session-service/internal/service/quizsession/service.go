package quizsession

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"time"

	"github.com/richardktran/realtime-quiz/pkg/cache/redis"
	"github.com/richardktran/realtime-quiz/pkg/events"
	messagebroker "github.com/richardktran/realtime-quiz/pkg/message-broker"
	"github.com/richardktran/realtime-quiz/pkg/topics"
	"github.com/richardktran/realtime-quiz/quiz-session-service/internal/repository"
	"github.com/richardktran/realtime-quiz/quiz-session-service/pkg/model"
)

var ErrNotFound = errors.New("not found")
var ErrUnauthorized = errors.New("unauthorized")
var ErrInvalidState = errors.New("invalid state")
var ErrAlreadyAnswered = errors.New("already answered")

type quizSessionRepository interface {
	CreateQuizSession(context.Context, *model.QuizSession) (*model.QuizSession, error)
	GetSessionById(context.Context, string) (*model.QuizSession, error)
	UpdateSession(context.Context, *model.QuizSession) error
	UpdateHost(context.Context, string, string) error
}

type quizBankGateway interface {
	GetRandomQuestions(ctx context.Context, count int32) ([]model.Question, error)
}

type Service struct {
	repo     quizSessionRepository
	producer messagebroker.Producer
	cache    *redis.Client
	quizBank quizBankGateway
}

func New(repo quizSessionRepository, producer messagebroker.Producer, cache *redis.Client, quizBank quizBankGateway) *Service {
	return &Service{
		repo:     repo,
		producer: producer,
		cache:    cache,
		quizBank: quizBank,
	}
}

func (s *Service) CreateQuizSession(ctx context.Context, data *model.QuizSession) (*model.QuizSession, error) {
	if data.Status == "" {
		data.Status = model.StatusWaiting
	}
	if data.QuestionIDs == nil {
		data.QuestionIDs = []string{}
	}

	// Persist immediately so join FK inserts succeed before the Kafka consumer runs.
	if _, err := s.repo.CreateQuizSession(ctx, data); err != nil {
		return nil, err
	}

	encoded, err := json.Marshal(data)
	if err != nil {
		return nil, err
	}

	if err := s.producer.Produce(ctx, topics.QuizSessionCreated, encoded); err != nil {
		return nil, err
	}

	state := &model.SessionState{
		ID:           data.ID,
		Name:         data.Name,
		HostID:       data.HostID,
		Status:       data.Status,
		Participants: map[string]*model.Participant{},
		Questions:    []model.Question{},
	}
	_ = s.saveState(ctx, state)

	return data, nil
}

func (s *Service) GetSessionById(ctx context.Context, id string) (*model.QuizSession, error) {
	if state, err := s.loadState(ctx, id); err == nil && state != nil {
		return sessionFromState(state), nil
	}

	session, err := s.repo.GetSessionById(ctx, id)
	if err != nil && errors.Is(err, repository.ErrNotFound) {
		return nil, ErrNotFound
	}
	return session, err
}

func (s *Service) JoinQuiz(ctx context.Context, quizSessionId, userId, name string) (*model.QuizSession, error) {
	session, err := s.GetSessionById(ctx, quizSessionId)
	if err != nil {
		return nil, err
	}

	state, err := s.loadOrInitState(ctx, session)
	if err != nil {
		return nil, err
	}

	if existing, ok := state.Participants[userId]; ok {
		// Reconnect: refresh name, keep score
		if name != "" {
			existing.Name = name
		}
	} else {
		state.Participants[userId] = &model.Participant{
			ID:    userId,
			Name:  name,
			Score: 0,
		}
	}

	if state.HostID == "" {
		state.HostID = userId
		session.HostID = userId
	}

	if err := s.saveState(ctx, state); err != nil {
		return nil, err
	}

	payload := events.UserJoined{
		SessionID: session.ID,
		UserID:    userId,
		Name:      name,
	}
	encoded, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}
	if err := s.producer.Produce(ctx, topics.UserJoinedQuiz, encoded); err != nil {
		return nil, err
	}

	return sessionFromState(state), nil
}

func (s *Service) StartSession(ctx context.Context, sessionID, userID string, questionCount int32) (*model.QuizSession, model.PublicQuestion, error) {
	locked, err := s.cache.SetNX(ctx, redis.SessionLockKey(sessionID, "start"), "1", 5*time.Second)
	if err != nil {
		return nil, model.PublicQuestion{}, err
	}
	if !locked {
		return nil, model.PublicQuestion{}, ErrInvalidState
	}
	defer s.cache.Del(ctx, redis.SessionLockKey(sessionID, "start"))

	session, err := s.GetSessionById(ctx, sessionID)
	if err != nil {
		return nil, model.PublicQuestion{}, err
	}

	state, err := s.loadOrInitState(ctx, session)
	if err != nil {
		return nil, model.PublicQuestion{}, err
	}

	if state.HostID != userID {
		return nil, model.PublicQuestion{}, ErrUnauthorized
	}

	// Idempotent start: if already in progress, re-emit the current question.
	if state.Status == model.StatusInProgress && len(state.Questions) > 0 {
		idx := state.CurrentQuestionIndex
		if idx < 0 || idx >= len(state.Questions) {
			idx = 0
		}
		publicQ := state.Questions[idx].ToPublic()
		payload := events.SessionStart{
			SessionID:     sessionID,
			QuestionIndex: idx,
			Question: events.PublicQuestion{
				ID:        publicQ.ID,
				Question:  publicQ.Question,
				Options:   publicQ.Options,
				TimeLimit: publicQ.TimeLimit,
			},
			Participants: toEventParticipants(participantsList(state)),
			HostID:       state.HostID,
		}
		if encoded, err := json.Marshal(payload); err == nil {
			_ = s.producer.Produce(ctx, topics.SessionStart, encoded)
		}
		return sessionFromState(state), publicQ, nil
	}

	if state.Status != model.StatusWaiting {
		return nil, model.PublicQuestion{}, ErrInvalidState
	}

	if questionCount <= 0 {
		questionCount = 5
	}

	questions, err := s.quizBank.GetRandomQuestions(ctx, questionCount)
	if err != nil {
		return nil, model.PublicQuestion{}, fmt.Errorf("fetch questions: %w", err)
	}
	if len(questions) == 0 {
		return nil, model.PublicQuestion{}, errors.New("no questions available")
	}

	ids := make([]string, len(questions))
	for i, q := range questions {
		ids[i] = q.ID
	}

	state.Questions = questions
	state.CurrentQuestionIndex = 0
	state.Status = model.StatusInProgress

	if err := s.saveState(ctx, state); err != nil {
		return nil, model.PublicQuestion{}, err
	}

	session.Status = model.StatusInProgress
	session.QuestionIDs = ids
	session.CurrentIndex = 0
	_ = s.repo.UpdateSession(ctx, session)

	publicQ := questions[0].ToPublic()
	participants := participantsList(state)

	payload := events.SessionStart{
		SessionID:     sessionID,
		QuestionIndex: 0,
		Question: events.PublicQuestion{
			ID:        publicQ.ID,
			Question:  publicQ.Question,
			Options:   publicQ.Options,
			TimeLimit: publicQ.TimeLimit,
		},
		Participants: toEventParticipants(participants),
		HostID:       state.HostID,
	}
	encoded, err := json.Marshal(payload)
	if err != nil {
		return nil, model.PublicQuestion{}, err
	}
	if err := s.producer.Produce(ctx, topics.SessionStart, encoded); err != nil {
		return nil, model.PublicQuestion{}, err
	}

	s.scheduleQuestionReveal(sessionID, publicQ.ID, state.CurrentQuestionIndex, publicQ.TimeLimit)

	return sessionFromState(state), publicQ, nil
}

func (s *Service) SubmitAnswer(ctx context.Context, sessionID, userID, questionID string, selectedOption, timeToAnswer int) error {
	state, err := s.loadState(ctx, sessionID)
	if err != nil {
		return ErrNotFound
	}
	if state.Status != model.StatusInProgress {
		return ErrInvalidState
	}

	participant, ok := state.Participants[userID]
	if !ok {
		return ErrNotFound
	}

	if state.CurrentQuestionIndex < 0 || state.CurrentQuestionIndex >= len(state.Questions) {
		return ErrInvalidState
	}
	question := state.Questions[state.CurrentQuestionIndex]
	if questionID != "" && question.ID != questionID {
		return ErrInvalidState
	}

	dedupeKey := redis.AnswerDedupKey(sessionID, question.ID, userID)
	okNX, err := s.cache.SetNX(ctx, dedupeKey, "1", 2*time.Hour)
	if err != nil {
		return err
	}
	if !okNX {
		return ErrAlreadyAnswered
	}

	answersKey := redis.QuestionAnswersKey(sessionID, question.ID)
	if err := s.cache.SAdd(ctx, answersKey, userID); err != nil {
		return err
	}
	_ = s.cache.Expire(ctx, answersKey, 2*time.Hour)

	payload := events.AnswerSubmitted{
		SessionID:      sessionID,
		UserID:         userID,
		Name:           participant.Name,
		QuestionID:     question.ID,
		SelectedOption: selectedOption,
		CorrectOption:  question.CorrectAnswer,
		TimeToAnswer:   timeToAnswer,
		QuestionIndex:  state.CurrentQuestionIndex,
	}
	encoded, err := json.Marshal(payload)
	if err != nil {
		return err
	}
	if err := s.producer.Produce(ctx, topics.AnswerSubmitted, encoded); err != nil {
		return err
	}

	submitted, err := s.cache.SCard(ctx, answersKey)
	if err != nil {
		return nil
	}
	if int(submitted) >= len(state.Participants) {
		_ = s.revealQuestion(ctx, sessionID, question.ID, state.CurrentQuestionIndex, question.CorrectAnswer, "all_submitted")
	}

	return nil
}

// scheduleQuestionReveal publishes question-result after the time limit if not already revealed.
func (s *Service) scheduleQuestionReveal(sessionID, questionID string, questionIndex, timeLimit int) {
	if timeLimit <= 0 {
		timeLimit = 15
	}
	go func() {
		time.Sleep(time.Duration(timeLimit) * time.Second)
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		state, err := s.loadState(ctx, sessionID)
		if err != nil {
			return
		}
		if state.Status != model.StatusInProgress {
			return
		}
		if state.CurrentQuestionIndex != questionIndex {
			return
		}
		if questionIndex < 0 || questionIndex >= len(state.Questions) {
			return
		}
		q := state.Questions[questionIndex]
		if q.ID != questionID {
			return
		}
		if err := s.revealQuestion(ctx, sessionID, questionID, questionIndex, q.CorrectAnswer, "timeout"); err != nil {
			log.Printf("reveal on timeout failed: %v", err)
		}
	}()
}

func (s *Service) revealQuestion(ctx context.Context, sessionID, questionID string, questionIndex, correctAnswer int, reason string) error {
	claimed, err := s.cache.SetNX(ctx, redis.QuestionRevealKey(sessionID, questionID), reason, 2*time.Hour)
	if err != nil {
		return err
	}
	if !claimed {
		return nil
	}

	payload := events.QuestionResult{
		SessionID:     sessionID,
		QuestionID:    questionID,
		QuestionIndex: questionIndex,
		CorrectAnswer: correctAnswer,
		Reason:        reason,
	}

	// Fan out immediately over Redis so clients are not stuck if the Kafka
	// consumer has not yet assigned the question-result topic.
	fanoutPayload := map[string]interface{}{
		"questionId":    questionID,
		"questionIndex": questionIndex,
		"correctAnswer": correctAnswer,
		"reason":        reason,
	}
	fanoutMsg, err := json.Marshal(events.FanoutMessage{
		SessionID: sessionID,
		Type:      "question_result",
		Payload:   fanoutPayload,
	})
	if err != nil {
		return err
	}
	if err := s.cache.Publish(ctx, topics.NotificationFanout, string(fanoutMsg)); err != nil {
		log.Printf("redis fanout question_result failed: %v", err)
	}

	encoded, err := json.Marshal(payload)
	if err != nil {
		return err
	}
	if err := s.producer.Produce(ctx, topics.QuestionResult, encoded); err != nil {
		log.Printf("kafka produce question-result failed: %v", err)
		return err
	}
	return nil
}

func (s *Service) NextQuestion(ctx context.Context, sessionID, userID string) (finished bool, question model.PublicQuestion, index int, err error) {
	locked, err := s.cache.SetNX(ctx, redis.SessionLockKey(sessionID, "next"), "1", 5*time.Second)
	if err != nil {
		return false, model.PublicQuestion{}, 0, err
	}
	if !locked {
		return false, model.PublicQuestion{}, 0, ErrInvalidState
	}
	defer s.cache.Del(ctx, redis.SessionLockKey(sessionID, "next"))

	state, err := s.loadState(ctx, sessionID)
	if err != nil {
		return false, model.PublicQuestion{}, 0, ErrNotFound
	}
	if state.HostID != userID {
		return false, model.PublicQuestion{}, 0, ErrUnauthorized
	}
	if state.Status != model.StatusInProgress {
		return false, model.PublicQuestion{}, 0, ErrInvalidState
	}

	nextIndex := state.CurrentQuestionIndex + 1
	if nextIndex >= len(state.Questions) {
		return true, model.PublicQuestion{}, nextIndex, s.EndSession(ctx, sessionID, userID)
	}

	state.CurrentQuestionIndex = nextIndex
	if err := s.saveState(ctx, state); err != nil {
		return false, model.PublicQuestion{}, 0, err
	}

	session, _ := s.repo.GetSessionById(ctx, sessionID)
	if session != nil {
		session.CurrentIndex = nextIndex
		_ = s.repo.UpdateSession(ctx, session)
	}

	publicQ := state.Questions[nextIndex].ToPublic()
	payload := events.ChangeQuestion{
		SessionID:     sessionID,
		QuestionIndex: nextIndex,
		Question: events.PublicQuestion{
			ID:        publicQ.ID,
			Question:  publicQ.Question,
			Options:   publicQ.Options,
			TimeLimit: publicQ.TimeLimit,
		},
	}
	encoded, err := json.Marshal(payload)
	if err != nil {
		return false, model.PublicQuestion{}, 0, err
	}
	if err := s.producer.Produce(ctx, topics.ChangeQuestion, encoded); err != nil {
		return false, model.PublicQuestion{}, 0, err
	}

	s.scheduleQuestionReveal(sessionID, publicQ.ID, nextIndex, publicQ.TimeLimit)

	return false, publicQ, nextIndex, nil
}

func (s *Service) EndSession(ctx context.Context, sessionID, userID string) error {
	state, err := s.loadState(ctx, sessionID)
	if err != nil {
		return ErrNotFound
	}
	if userID != "" && state.HostID != userID {
		return ErrUnauthorized
	}

	// Sync final scores from the leaderboard sorted set when available.
	if rankings, err := s.cache.ZRevRangeWithScores(ctx, redis.LeaderboardKey(sessionID), 0, -1); err == nil {
		for _, z := range rankings {
			member, _ := z.Member.(string)
			if p, ok := state.Participants[member]; ok {
				p.Score = int(z.Score)
			}
		}
	}

	state.Status = model.StatusFinished
	if err := s.saveState(ctx, state); err != nil {
		return err
	}

	session, repoErr := s.repo.GetSessionById(ctx, sessionID)
	if repoErr == nil && session != nil {
		session.Status = model.StatusFinished
		_ = s.repo.UpdateSession(ctx, session)
	}

	payload := events.SessionEnd{
		SessionID:    sessionID,
		Participants: toEventParticipants(participantsList(state)),
	}
	encoded, err := json.Marshal(payload)
	if err != nil {
		return err
	}
	return s.producer.Produce(ctx, topics.SessionEnd, encoded)
}

func (s *Service) ReassignHost(ctx context.Context, sessionID, leavingUserID string) (string, error) {
	state, err := s.loadState(ctx, sessionID)
	if err != nil {
		return "", ErrNotFound
	}

	delete(state.Participants, leavingUserID)

	if state.HostID == leavingUserID {
		state.HostID = ""
		for id := range state.Participants {
			state.HostID = id
			break
		}
	}

	if err := s.saveState(ctx, state); err != nil {
		return "", err
	}
	if state.HostID != "" {
		_ = s.repo.UpdateHost(ctx, sessionID, state.HostID)
	}
	return state.HostID, nil
}

func (s *Service) loadState(ctx context.Context, sessionID string) (*model.SessionState, error) {
	raw, err := s.cache.Get(ctx, redis.SessionKey(sessionID))
	if err != nil {
		return nil, err
	}
	var state model.SessionState
	if err := json.Unmarshal([]byte(raw), &state); err != nil {
		return nil, err
	}
	if state.Participants == nil {
		state.Participants = map[string]*model.Participant{}
	}
	return &state, nil
}

func (s *Service) saveState(ctx context.Context, state *model.SessionState) error {
	encoded, err := json.Marshal(state)
	if err != nil {
		return err
	}
	return s.cache.Set(ctx, redis.SessionKey(state.ID), encoded, 24*time.Hour)
}

func (s *Service) loadOrInitState(ctx context.Context, session *model.QuizSession) (*model.SessionState, error) {
	if state, err := s.loadState(ctx, session.ID); err == nil {
		return state, nil
	}

	participants := map[string]*model.Participant{}
	for i := range session.Participants {
		p := session.Participants[i]
		participants[p.ID] = &p
	}

	state := &model.SessionState{
		ID:                   session.ID,
		Name:                 session.Name,
		HostID:               session.HostID,
		Status:               session.Status,
		CurrentQuestionIndex: session.CurrentIndex,
		Participants:         participants,
		Questions:            []model.Question{},
	}
	if err := s.saveState(ctx, state); err != nil {
		return nil, err
	}
	return state, nil
}

func sessionFromState(state *model.SessionState) *model.QuizSession {
	ids := make([]string, 0, len(state.Questions))
	for _, q := range state.Questions {
		ids = append(ids, q.ID)
	}
	return &model.QuizSession{
		ID:           state.ID,
		Name:         state.Name,
		HostID:       state.HostID,
		Status:       state.Status,
		CurrentIndex: state.CurrentQuestionIndex,
		QuestionIDs:  ids,
		Participants: participantsList(state),
	}
}

func participantsList(state *model.SessionState) []model.Participant {
	out := make([]model.Participant, 0, len(state.Participants))
	for _, p := range state.Participants {
		if p != nil {
			out = append(out, *p)
		}
	}
	return out
}

func toEventParticipants(ps []model.Participant) []events.Participant {
	out := make([]events.Participant, 0, len(ps))
	for _, p := range ps {
		out = append(out, events.Participant{ID: p.ID, Name: p.Name, Score: p.Score})
	}
	return out
}
