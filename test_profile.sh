#!/bin/bash
echo "Testing profile functionality with bin/magellai ask command"
echo

echo "1. Testing with the creative profile (should use anthropic)"
bin/magellai ask -c magellai.config.yaml --profile=creative "Are you a Claude AI model?"

echo
echo "2. Testing with the quality profile (should use openai)"
bin/magellai ask -c magellai.config.yaml --profile=quality "Are you a GPT model?"

echo
echo "Test completed."