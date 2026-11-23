FROM node:22-bookworm-slim@sha256:e9f565d32859f0cbfb52125eafe8be39390f24dac302dc3cdf566e8ba3ab21ee
WORKDIR /opt/reference
COPY package.json package-lock.json ./
RUN npm ci --ignore-scripts --omit=dev --no-audit --no-fund && npm cache clean --force
COPY render.cjs ./
USER 65534:65534
ENTRYPOINT ["node", "render.cjs"]
