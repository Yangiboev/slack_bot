package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"

	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
	"github.com/shomali11/slacker"
	"github.com/spf13/viper"
)

func init() {
	// Load environment variables from .env file
	err := godotenv.Load()
	if err != nil {
		log.Fatalf("Error loading .env file: %s", err)
	}

	// Set up Viper to read environment variables
	viper.AutomaticEnv()
	viper.SetDefault("DB_HOST", "localhost")
	viper.SetDefault("DB_PORT", "5432")
	viper.SetDefault("DB_USER", "postgres")
	viper.SetDefault("DB_PASSWORD", "")
	viper.SetDefault("DB_NAME", "my_database")
	viper.SetDefault("SLACK_BOT_TOKEN", "yourtoken")
	viper.SetDefault("SLACK_APP_TOKEN", "yourtoken")
}

func printCommandEvents(analyticsChannel <-chan *slacker.CommandEvent) {
	for event := range analyticsChannel {
		log.Printf("Params: %+v", event.Parameters)
		log.Printf("Time: %s\n", event.Timestamp.UTC().String())
	}
}

func main() {
	botToken := viper.GetString("SLACK_BOT_TOKEN")
	appToken := viper.GetString("SLACK_APP_TOKEN")

	bot := slacker.NewClient(botToken, appToken)

	go printCommandEvents(bot.CommandEvents())

	bot.Command("ping", &slacker.CommandDefinition{
		Handler: func(botCtx slacker.BotContext, request slacker.Request, response slacker.ResponseWriter) {
			response.Reply("pong")
		},
	})

	bot.Command("disable game {game}", &slacker.CommandDefinition{
		Description: "Disable the game given!",
		Examples:    []string{"disable game ABC"},
		Handler: func(botCtx slacker.BotContext, request slacker.Request, response slacker.ResponseWriter) {
			game := request.Param("game")
			if err := executeDBCommand(game); err != nil {
				response.ReportError(fmt.Errorf("failed! %s", game))
			} else {
				response.Reply("Successfully updated!")
			}
		},
	})

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	err := bot.Listen(ctx)
	if err != nil {
		log.Fatal(err)
	}
}

func executeDBCommand(name string) error {
	dbHost := viper.GetString("DB_HOST")
	dbPort := viper.GetString("DB_PORT")
	dbUser := viper.GetString("DB_USER")
	dbPassword := viper.GetString("DB_PASSWORD")
	dbName := viper.GetString("DB_NAME")

	dsn := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable", dbHost, dbPort, dbUser, dbPassword, dbName)

	db, err := sql.Open("postgres", dsn)
	if err != nil {
		return fmt.Errorf("failed to open connection: %w", err)
	}
	defer db.Close()

	res, err := db.Exec("UPDATE games SET disabled=true WHERE name= $1", name)
	if err != nil {
		return fmt.Errorf("failed to execute: %w", err)
	}

	rowsAffected, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get efected: %w", err)
	}
	fmt.Printf("%d rows affected.\n", rowsAffected)
	return nil
}
