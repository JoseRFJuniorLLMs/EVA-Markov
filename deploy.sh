#!/bin/bash

# Deploy EVA-Markov to Google Cloud Run

set -e

PROJECT_ID="your-gcp-project-id"
REGION="southamerica-east1"
SERVICE_NAME="eva-markov"

echo "🚀 Deploying EVA-Markov to Cloud Run..."

# Build image
echo "📦 Building Docker image..."
gcloud builds submit --tag gcr.io/$PROJECT_ID/$SERVICE_NAME

# Deploy to Cloud Run
echo "☁️ Deploying to Cloud Run..."
gcloud run deploy $SERVICE_NAME \
  --image gcr.io/$PROJECT_ID/$SERVICE_NAME \
  --platform managed \
  --region $REGION \
  --allow-unauthenticated \
  --set-env-vars="ENV=production" \
  --set-secrets="GOOGLE_API_KEY=GOOGLE_API_KEY:latest,DATABASE_URL=DATABASE_URL:latest" \
  --memory 512Mi \
  --cpu 1 \
  --timeout 3600 \
  --max-instances 1 \
  --min-instances 0

echo "✅ Deploy concluído!"
echo "📊 Service URL:"
gcloud run services describe $SERVICE_NAME --region $REGION --format 'value(status.url)'
