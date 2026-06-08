package main

import (
	"log"
	"net/http"
	"os"
)

func main() {
	token := os.Getenv("LIFEOPS_WEBHOOK_TOKEN")
	if token == "" {
		token = "dev-token"
	}
	addr := os.Getenv("LIFEOPS_ADDR")
	if addr == "" {
		addr = ":8080"
	}
	dbPath := os.Getenv("LIFEOPS_DB_PATH")
	if dbPath == "" {
		dbPath = "lifeops.db"
	}

	store, err := NewStore(dbPath)
	if err != nil {
		log.Fatalf("failed to open database: %v", err)
	}
	defer store.Close()

	// Initialize LLM client from config or env vars
	var butler *ButlerAgent
	aiConfig, _ := store.GetAIConfig()

	endpoint := os.Getenv("LIFEOPS_AI_ENDPOINT")
	apiKey := os.Getenv("LIFEOPS_AI_API_KEY")
	model := os.Getenv("LIFEOPS_AI_MODEL")

	if aiConfig != nil {
		if endpoint == "" {
			endpoint = aiConfig.Endpoint
		}
		if apiKey == "" {
			apiKey = aiConfig.APIKey
		}
		if model == "" {
			model = aiConfig.Model
		}
	}

	if endpoint != "" && apiKey != "" {
		llm := NewLLMClient(LLMConfig{
			Endpoint:  endpoint,
			APIKey:    apiKey,
			Model:     model,
			MaxTokens: 2048,
		})

		var router *SkillRouter
		skillsPath := os.Getenv("LIFESTYLE_SKILLS_PATH")
		if skillsPath != "" {
			reg, err := NewSkillRegistry(skillsPath)
			if err != nil {
				log.Printf("WARNING: failed to load lifestyle skills from %s: %v", skillsPath, err)
			} else if reg.Available() {
				router = NewSkillRouter(reg, llm)
				log.Printf("Lifestyle skills loaded from %s (%d lenses, domains: %v)", skillsPath, len(reg.ListLenses("")), reg.Domains())
			}
		}

		butler = NewButlerAgent(store, llm, router)
		log.Printf("AI butler initialized (model: %s, endpoint: %s)", model, endpoint)
	} else {
		log.Printf("WARNING: AI not configured. Set LIFEOPS_AI_ENDPOINT and LIFEOPS_AI_API_KEY or configure via /api/config/ai")
	}

	log.Printf("LifeOps API listening on %s (db: %s)", addr, dbPath)
	if err := http.ListenAndServe(addr, NewServer(token, store, butler)); err != nil {
		log.Fatal(err)
	}
}
