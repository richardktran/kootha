# Build from repository root:
#   docker build -f web/Dockerfile -t kootha/web:latest .

FROM oven/bun:1 AS builder

WORKDIR /app
COPY web/package.json web/bun.lockb ./
RUN bun install --frozen-lockfile
COPY web/ .

# Placeholders are replaced at container start so the same image works in any environment.
ENV NEXT_PUBLIC_API_URL=__NEXT_PUBLIC_API_URL__
ENV NEXT_PUBLIC_WS_URL=__NEXT_PUBLIC_WS_URL__
RUN bun run build

FROM node:22-alpine AS runner

WORKDIR /app
ENV NODE_ENV=production
ENV PORT=3000

COPY --from=builder /app/public ./public
COPY --from=builder /app/.next/standalone ./
COPY --from=builder /app/.next/static ./.next/static
COPY web/entrypoint.sh ./entrypoint.sh
RUN chmod +x ./entrypoint.sh

EXPOSE 3000
ENTRYPOINT ["./entrypoint.sh"]
