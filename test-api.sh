#!/bin/bash

# Colors for output
GREEN='\033[0;32m'
RED='\033[0;31m'
NC='\033[0m' # No Color

echo "🚀 Cluster-Imager API Test Script"
echo "================================="

# Check if server is running
if ! curl -s http://localhost:8080 > /dev/null 2>&1; then
    echo -e "${RED}❌ Server is not running on localhost:8080${NC}"
    echo "Start the server with: make run"
    exit 1
fi

echo -e "${GREEN}✅ Server is running${NC}"

# Create test directory
mkdir -p test-images
cd test-images

# Download sample image if not exists
if [ ! -f "sample.jpg" ]; then
    echo "📥 Downloading sample image..."
    curl -s -o sample.jpg https://picsum.photos/800/600
    echo -e "${GREEN}✅ Sample image downloaded${NC}"
fi

# Test 1: Resize
echo -e "\n📐 Test 1: Resize to 400x300"
curl -s -X POST \
    -F "image=@sample.jpg" \
    "http://localhost:8080/resize?width=400&height=300" \
    --output resized-400x300.jpg

if [ -f "resized-400x300.jpg" ]; then
    SIZE=$(identify -format "%wx%h" resized-400x300.jpg 2>/dev/null || echo "unknown")
    echo -e "${GREEN}✅ Resized image created: $SIZE${NC}"
else
    echo -e "${RED}❌ Resize failed${NC}"
fi

# Test 2: Crop
echo -e "\n✂️  Test 2: Crop 200x200 from (100,100)"
curl -s -X POST \
    -F "image=@sample.jpg" \
    "http://localhost:8080/crop?x=100&y=100&width=200&height=200" \
    --output cropped-200x200.jpg

if [ -f "cropped-200x200.jpg" ]; then
    SIZE=$(identify -format "%wx%h" cropped-200x200.jpg 2>/dev/null || echo "unknown")
    echo -e "${GREEN}✅ Cropped image created: $SIZE${NC}"
else
    echo -e "${RED}❌ Crop failed${NC}"
fi

# Test 3: Error handling - negative dimensions
echo -e "\n🚫 Test 3: Error handling (negative dimensions)"
RESPONSE=$(curl -s -X POST \
    -F "image=@sample.jpg" \
    "http://localhost:8080/resize?width=-100&height=200" \
    -w "\nHTTP_CODE:%{http_code}")

HTTP_CODE=$(echo "$RESPONSE" | grep "HTTP_CODE:" | cut -d: -f2)
if [ "$HTTP_CODE" = "400" ]; then
    echo -e "${GREEN}✅ Correctly rejected negative dimensions (HTTP 400)${NC}"
else
    echo -e "${RED}❌ Expected HTTP 400, got $HTTP_CODE${NC}"
fi

# Test 4: Error handling - crop outside bounds
echo -e "\n🚫 Test 4: Error handling (crop outside bounds)"
RESPONSE=$(curl -s -X POST \
    -F "image=@sample.jpg" \
    "http://localhost:8080/crop?x=700&y=500&width=200&height=200" \
    -w "\nHTTP_CODE:%{http_code}")

HTTP_CODE=$(echo "$RESPONSE" | grep "HTTP_CODE:" | cut -d: -f2)
if [ "$HTTP_CODE" = "400" ]; then
    echo -e "${GREEN}✅ Correctly rejected out-of-bounds crop (HTTP 400)${NC}"
else
    echo -e "${RED}❌ Expected HTTP 400, got $HTTP_CODE${NC}"
fi

# Test 5: Request tracking
echo -e "\n🔍 Test 5: Request ID tracking"
REQUEST_ID=$(curl -s -X POST \
    -F "image=@sample.jpg" \
    "http://localhost:8080/resize?width=100&height=100" \
    -D - \
    --output /dev/null | grep -i "x-request-id" | cut -d' ' -f2 | tr -d '\r')

if [ ! -z "$REQUEST_ID" ]; then
    echo -e "${GREEN}✅ Request ID: $REQUEST_ID${NC}"
else
    echo -e "${RED}❌ No Request ID found${NC}"
fi

echo -e "\n✨ Testing complete!"
echo "📁 Results saved in: $(pwd)"
echo -e "\nView images with:"
echo "  open *.jpg  # macOS"
echo "  xdg-open *.jpg  # Linux"