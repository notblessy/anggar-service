package model

import (
	"context"
	"fmt"
)

type RecognizerRepository interface {
	RecognizeTransaction(ctx context.Context, prompt, text string) (Transaction, error)
}

func SystemPrompt(meID, sharedID string) string {
	return fmt.Sprintf(`
		You are a finance message parser. Given a short, natural language message like "makan ayam 50000" or "uang freelance 200000", respond with a structured JSON that fits this format:
		{
			"id": "string",                       // generate as ulid
			"description": "string",               // short description of the transaction
			"amount": number,                      // numeric amount in IDR
			"transaction_type": "INCOME" | "EXPENSE", // determine based on message
			"wallet_name": "string",               // if not mentioned, return "default"
			"user_id": "string"                      // if not mentioned, return "self"
			"is_shared": true | false,
			"category": "string"                  // detect category from description
			"spent_at": "2023-10-01T00:00:00Z" // if not mentioned, return current time
		}

		if message contains (name <amount>, name <amount>), return is_shared: true,
		you can assume the first name is me, and the second name is shared.
		also assume the me name with %s ID and shared name with %s ID
		amount might not be percentage, if amount has %%, then it's percentage and calculate the shared amount based on amount transaction with the percentage.
		For example:
		- "makan ayam 50000 (shelly 50%%, blessy 50%%)" will return 25000 for me and 25000 for shared
		if the message doesn't contain percentage, then just split with the exact amount, and find the percentage based on the amount.
		also return json with transaction_shares array of objects with the following format:
		{
			"id": "string", // generate as ulid
			"transaction_id": "string", // id of the transaction
			"user_id": "",
			"percentage": number, // percentage of the share
			"amount": number // amount of the share
		}
		Assume:
		- It's always an expense transaction
		- Detect which description and which is amount

		Only respond with the JSON object. No explanation, no extra text.
	`, meID, sharedID)
}
