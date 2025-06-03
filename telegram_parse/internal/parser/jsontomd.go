package parser

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"telegram_parse/internal/telegram"
)

// JSONToMarkdown converts Telegram JSON export to clean Markdown
type JSONToMarkdown struct {
	// Options for formatting
	includeMetadata bool
	includeMedia    bool
	dateFormat      string
}

// NewJSONToMarkdown creates a new JSON to Markdown converter
func NewJSONToMarkdown() *JSONToMarkdown {
	return &JSONToMarkdown{
		includeMetadata: true,
		includeMedia:    true,
		dateFormat:      "2006-01-02 15:04:05",
	}
}

// ConvertFile converts JSON file to Markdown
func (p *JSONToMarkdown) ConvertFile(inputPath, outputPath string) error {
	// Open input file
	file, err := os.Open(inputPath)
	if err != nil {
		return fmt.Errorf("failed to open input file: %w", err)
	}
	defer file.Close()

	// Parse JSON
	var export telegram.Export
	decoder := json.NewDecoder(file)
	err = decoder.Decode(&export)
	if err != nil {
		return fmt.Errorf("failed to parse JSON: %w", err)
	}

	// Convert to Markdown
	markdown := p.exportToMarkdown(&export)

	// Write to output file
	outFile, err := os.Create(outputPath)
	if err != nil {
		return fmt.Errorf("failed to create output file: %w", err)
	}
	defer outFile.Close()

	_, err = outFile.WriteString(markdown)
	if err != nil {
		// Clean up failed output file
		os.Remove(outputPath)
		return fmt.Errorf("failed to write output: %w", err)
	}

	return nil
}

// exportToMarkdown converts Export struct to Markdown
func (p *JSONToMarkdown) exportToMarkdown(export *telegram.Export) string {
	var result strings.Builder

	// Add header with chat information
	if p.includeMetadata {
		result.WriteString(fmt.Sprintf("# %s\n\n", export.Name))
		result.WriteString(fmt.Sprintf("**Type:** %s  \n", export.Type))
		if export.ID != 0 {
			result.WriteString(fmt.Sprintf("**ID:** %d  \n", export.ID))
		}
		result.WriteString(fmt.Sprintf("**Messages:** %d  \n\n", len(export.Messages)))
		result.WriteString("---\n\n")
	}

	// Process messages
	for _, message := range export.Messages {
		messageMarkdown := p.messageToMarkdown(&message)
		if messageMarkdown != "" {
			result.WriteString(messageMarkdown)
			result.WriteString("\n")
		}
	}

	return result.String()
}

// messageToMarkdown converts a single message to Markdown
func (p *JSONToMarkdown) messageToMarkdown(msg *telegram.Message) string {
	var result strings.Builder

	// Skip service messages without meaningful content
	if msg.Type == "service" && msg.Text == nil && msg.Action == "" {
		return ""
	}

	// Add message header
	result.WriteString("## ")

	// Add date
	if msg.Date != "" {
		parsedDate := p.parseDate(msg.Date)
		result.WriteString(parsedDate.Format(p.dateFormat))
	}

	// Add sender
	if msg.From != "" {
		result.WriteString(fmt.Sprintf(" - %s", msg.From))
	}

	result.WriteString("\n\n")

	// Handle different message types
	switch msg.Type {
	case "message":
		p.processRegularMessage(msg, &result)
	case "service":
		p.processServiceMessage(msg, &result)
	default:
		p.processRegularMessage(msg, &result)
	}

	// Add forwarded information
	if msg.ForwardedFrom != "" {
		result.WriteString(fmt.Sprintf("\n*Forwarded from: %s*\n", msg.ForwardedFrom))
	}

	// Add reply information
	if msg.ReplyToMessageID != 0 {
		result.WriteString(fmt.Sprintf("\n*Reply to message ID: %d*\n", msg.ReplyToMessageID))
	}

	// Add via bot information
	if msg.ViaBot != "" {
		result.WriteString(fmt.Sprintf("\n*Via bot: %s*\n", msg.ViaBot))
	}

	result.WriteString("\n---\n\n")
	return result.String()
}

// processRegularMessage processes regular text messages
func (p *JSONToMarkdown) processRegularMessage(msg *telegram.Message, result *strings.Builder) {
	// Process text content
	if msg.Text != nil {
		textContent := p.extractTextContent(msg.Text, msg.TextEntities)
		if textContent != "" {
			result.WriteString(textContent)
			result.WriteString("\n\n")
		}
	}

	// Process media
	if p.includeMedia {
		p.processMedia(msg, result)
	}

	// Process polls
	if msg.Poll != nil {
		p.processPoll(msg.Poll, result)
	}

	// Process contact
	if msg.ContactInformation != nil {
		p.processContact(msg.ContactInformation, result)
	}

	// Process location
	if msg.LocationInformation != nil {
		p.processLocation(msg.LocationInformation, result)
	}
}

// processServiceMessage processes service messages (join, leave, etc.)
func (p *JSONToMarkdown) processServiceMessage(msg *telegram.Message, result *strings.Builder) {
	if msg.Action != "" {
		result.WriteString(fmt.Sprintf("*%s*", msg.Action))

		if msg.Actor != "" {
			result.WriteString(fmt.Sprintf(" by %s", msg.Actor))
		}

		if len(msg.Members) > 0 {
			result.WriteString(fmt.Sprintf(" - Members: %s", strings.Join(msg.Members, ", ")))
		}

		if msg.Inviter != "" {
			result.WriteString(fmt.Sprintf(" - Invited by: %s", msg.Inviter))
		}

		if msg.Title != "" {
			result.WriteString(fmt.Sprintf(" - Title: %s", msg.Title))
		}

		result.WriteString("\n\n")
	}
}

// extractTextContent processes text content with entities
func (p *JSONToMarkdown) extractTextContent(text interface{}, entities []telegram.TextEntity) string {
	switch t := text.(type) {
	case string:
		return p.cleanText(t)
	case []interface{}:
		return p.processTextEntities(t, entities)
	default:
		return ""
	}
}

// processTextEntities processes text entities and applies formatting
func (p *JSONToMarkdown) processTextEntities(textArray []interface{}, entities []telegram.TextEntity) string {
	var result strings.Builder

	for _, item := range textArray {
		switch v := item.(type) {
		case string:
			result.WriteString(p.cleanText(v))
		case map[string]interface{}:
			if textVal, ok := v["text"].(string); ok {
				if typeVal, ok := v["type"].(string); ok {
					formatted := p.formatText(textVal, typeVal, v)
					result.WriteString(formatted)
				} else {
					result.WriteString(p.cleanText(textVal))
				}
			}
		}
	}

	return result.String()
}

// formatText applies Markdown formatting based on entity type
func (p *JSONToMarkdown) formatText(text, entityType string, entity map[string]interface{}) string {
	cleanedText := p.cleanText(text)

	switch entityType {
	case "bold":
		return fmt.Sprintf("**%s**", cleanedText)
	case "italic":
		return fmt.Sprintf("*%s*", cleanedText)
	case "code":
		return fmt.Sprintf("`%s`", cleanedText)
	case "pre":
		return fmt.Sprintf("```\n%s\n```", cleanedText)
	case "text_link":
		if href, ok := entity["href"].(string); ok {
			return fmt.Sprintf("[%s](%s)", cleanedText, href)
		}
		return cleanedText
	case "mention":
		return fmt.Sprintf("@%s", cleanedText)
	case "hashtag":
		return fmt.Sprintf("#%s", cleanedText)
	case "strikethrough":
		return fmt.Sprintf("~~%s~~", cleanedText)
	case "underline":
		return fmt.Sprintf("__%s__", cleanedText)
	case "spoiler":
		return fmt.Sprintf("||%s||", cleanedText)
	default:
		return cleanedText
	}
}

// processMedia adds media information to markdown
func (p *JSONToMarkdown) processMedia(msg *telegram.Message, result *strings.Builder) {
	if msg.Photo != "" {
		result.WriteString(fmt.Sprintf("📷 **Photo:** %s", filepath.Base(msg.Photo)))
		if msg.Width > 0 && msg.Height > 0 {
			result.WriteString(fmt.Sprintf(" (%dx%d)", msg.Width, msg.Height))
		}
		result.WriteString("\n\n")
	}

	if msg.File != "" {
		result.WriteString(fmt.Sprintf("📎 **File:** %s", filepath.Base(msg.File)))
		if msg.MimeType != "" {
			result.WriteString(fmt.Sprintf(" (%s)", msg.MimeType))
		}
		if msg.Duration > 0 {
			result.WriteString(fmt.Sprintf(" - Duration: %d seconds", msg.Duration))
		}
		result.WriteString("\n\n")
	}

	if msg.MediaType != "" && msg.MediaType != "photo" {
		result.WriteString(fmt.Sprintf("🎬 **Media Type:** %s\n\n", msg.MediaType))
	}
}

// processPoll adds poll information to markdown
func (p *JSONToMarkdown) processPoll(poll *telegram.Poll, result *strings.Builder) {
	result.WriteString("📊 **Poll**\n\n")
	result.WriteString(fmt.Sprintf("**Question:** %s\n\n", poll.Question))

	if len(poll.Answers) > 0 {
		result.WriteString("**Options:**\n")
		for _, answer := range poll.Answers {
			marker := "☐"
			if answer.Chosen {
				marker = "☑"
			}
			result.WriteString(fmt.Sprintf("- %s %s (%d votes)\n", marker, answer.Text, answer.Voters))
		}
		result.WriteString("\n")
	}

	if poll.Closed {
		result.WriteString("*Poll is closed*\n")
	}

	result.WriteString(fmt.Sprintf("**Total voters:** %d\n\n", poll.TotalVoters))
}

// processContact adds contact information to markdown
func (p *JSONToMarkdown) processContact(contact *telegram.Contact, result *strings.Builder) {
	result.WriteString("📞 **Contact**\n\n")

	if contact.FirstName != "" || contact.LastName != "" {
		result.WriteString(fmt.Sprintf("**Name:** %s %s\n", contact.FirstName, contact.LastName))
	}

	if contact.PhoneNumber != "" {
		result.WriteString(fmt.Sprintf("**Phone:** %s\n", contact.PhoneNumber))
	}

	if contact.UserID != 0 {
		result.WriteString(fmt.Sprintf("**User ID:** %d\n", contact.UserID))
	}

	result.WriteString("\n")
}

// processLocation adds location information to markdown
func (p *JSONToMarkdown) processLocation(location *telegram.Location, result *strings.Builder) {
	result.WriteString("📍 **Location**\n\n")
	result.WriteString(fmt.Sprintf("**Coordinates:** %.6f, %.6f\n", location.Latitude, location.Longitude))
	result.WriteString(fmt.Sprintf("**Map Link:** https://maps.google.com/?q=%.6f,%.6f\n\n", location.Latitude, location.Longitude))
}

// parseDate parses Telegram date format
func (p *JSONToMarkdown) parseDate(dateStr string) time.Time {
	// Try parsing standard ISO format first
	if t, err := time.Parse("2006-01-02T15:04:05", dateStr); err == nil {
		return t
	}

	// Try parsing without timezone
	if t, err := time.Parse("2006-01-02 15:04:05", dateStr); err == nil {
		return t
	}

	// Return current time if parsing fails
	return time.Now()
}

// cleanText cleans text content, removing unwanted characters
func (p *JSONToMarkdown) cleanText(text string) string {
	// Remove excessive whitespace
	cleaned := regexp.MustCompile(`\s+`).ReplaceAllString(text, " ")
	cleaned = strings.TrimSpace(cleaned)

	// Escape markdown special characters that shouldn't be interpreted
	specialChars := []string{
		"\\", "`", "*", "_", "{", "}", "[", "]", "(", ")",
		"#", "+", "-", ".", "!", "|",
	}

	for _, char := range specialChars {
		cleaned = strings.ReplaceAll(cleaned, char, "\\"+char)
	}

	return cleaned
}
