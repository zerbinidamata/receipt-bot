#!/bin/bash
set -e

# Recipe Bot Deployment Script
# Builds images locally and deploys to Cloud Run via Terraform
# This avoids Cloud Build costs

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

echo -e "${GREEN}=== Recipe Bot Deployment ===${NC}"

# Check required tools
command -v docker >/dev/null 2>&1 || { echo -e "${RED}Docker is required but not installed.${NC}" >&2; exit 1; }
command -v gcloud >/dev/null 2>&1 || { echo -e "${RED}gcloud CLI is required but not installed.${NC}" >&2; exit 1; }
command -v terraform >/dev/null 2>&1 || { echo -e "${RED}Terraform is required but not installed.${NC}" >&2; exit 1; }

# Get project ID
PROJECT_ID=$(gcloud config get-value project 2>/dev/null)
if [ -z "$PROJECT_ID" ]; then
    echo -e "${RED}No GCP project configured. Run: gcloud config set project YOUR_PROJECT_ID${NC}"
    exit 1
fi

echo -e "${YELLOW}Project: ${PROJECT_ID}${NC}"

# Set variables
REGION=${REGION:-us-central1}
BOT_IMAGE="gcr.io/${PROJECT_ID}/recipe-bot:latest"
SCRAPER_IMAGE="gcr.io/${PROJECT_ID}/recipe-bot-scraper:latest"

# Navigate to project root
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(dirname "$SCRIPT_DIR")"
cd "$PROJECT_ROOT"

echo -e "${GREEN}Step 1: Configure Docker for GCR${NC}"
gcloud auth configure-docker gcr.io --quiet

echo -e "${GREEN}Step 2: Build Go Bot Image${NC}"
docker build -t "$BOT_IMAGE" -f Dockerfile .

echo -e "${GREEN}Step 3: Build Python Scraper Image${NC}"
docker build -t "$SCRAPER_IMAGE" -f python-service/Dockerfile python-service/

echo -e "${GREEN}Step 4: Push Images to GCR${NC}"
docker push "$BOT_IMAGE"
docker push "$SCRAPER_IMAGE"

echo -e "${GREEN}Step 5: Deploy with Terraform${NC}"
cd terraform

# Check if terraform.tfvars exists
if [ ! -f "terraform.tfvars" ]; then
    echo -e "${YELLOW}terraform.tfvars not found. Creating from example...${NC}"
    cp terraform.tfvars.example terraform.tfvars
    echo -e "${RED}Please edit terraform/terraform.tfvars with your API keys and run again.${NC}"
    exit 1
fi

# Update image references in tfvars
sed -i.bak "s|bot_image.*=.*|bot_image     = \"${BOT_IMAGE}\"|" terraform.tfvars
sed -i.bak "s|scraper_image.*=.*|scraper_image = \"${SCRAPER_IMAGE}\"|" terraform.tfvars
rm -f terraform.tfvars.bak

# Initialize and apply Terraform
terraform init
terraform apply -auto-approve

echo -e "${GREEN}=== Deployment Complete ===${NC}"
echo ""
terraform output

echo ""
echo -e "${YELLOW}To set up Telegram webhook, run:${NC}"
echo "curl 'https://api.telegram.org/bot<YOUR_TOKEN>/setWebhook?url=<BOT_URL>/webhook'"
