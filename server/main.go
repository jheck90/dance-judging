package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/joho/godotenv"
	"github.com/gin-gonic/gin"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/option"
	"google.golang.org/api/sheets/v4"
)

var (
	googleSheetID string
)

func main() {
		// Load environment variables from .envrc
		if err := godotenv.Load(".envrc"); err != nil {
			log.Fatal("Error loading .envrc file")
		}
	
		// Retrieve the Google Sheet ID from the environment variables
		googleSheetID = os.Getenv("GOOGLE_SHEET_ID")
		if googleSheetID == "" {
			log.Fatal("GOOGLE_SHEET_ID is not set in the .envrc file")
		}
	router := gin.Default()

	// router.GET("/google-sheets", handleGoogleSheets)
	router.StaticFS("/", http.Dir("client/build"))


	log.Fatal(router.Run(":6969"))
}

func handleGoogleSheets(c *gin.Context) {
	ctx := context.Background()
	b, err := os.ReadFile("credentials.json")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Unable to read client secret file"})
		return
	}

	config, err := google.ConfigFromJSON(b, "https://www.googleapis.com/auth/spreadsheets.readonly")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Unable to parse client secret file to config"})
		return
	}
	client := getClient(config)

	srv, err := sheets.NewService(ctx, option.WithHTTPClient(client))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Unable to retrieve Sheets client"})
		return
	}

	spreadsheetID := "12l7AU9V80NbvSVgqxHfA_Fp1yEUe22qEU66g3UDCni8"
	readRange := "2022_State!A6:H"
	resp, err := srv.Spreadsheets.Values.Get(spreadsheetID, readRange).Do()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Unable to retrieve data from sheet: %v", err)})
		return
	}

	if len(resp.Values) == 0 {
		c.JSON(http.StatusOK, gin.H{"message": "No data found"})
	} else {
		c.JSON(http.StatusOK, gin.H{"data": resp.Values})
	}
}

func getClient(config *oauth2.Config) *http.Client {
	tokFile := "token.json"
	tok, err := tokenFromFile(tokFile)
	if err != nil {
		tok = getTokenFromWeb(config)
		saveToken(tokFile, tok)
	}
	return config.Client(context.Background(), tok)
}

func getTokenFromWeb(config *oauth2.Config) *oauth2.Token {
	authURL := config.AuthCodeURL("state-token", oauth2.AccessTypeOffline)
	fmt.Printf("Go to the following link in your browser then type the "+
		"authorization code: \n%v\n", authURL)

	var authCode string
	if _, err := fmt.Scan(&authCode); err != nil {
		log.Fatalf("Unable to read authorization code: %v", err)
	}

	tok, err := config.Exchange(context.TODO(), authCode)
	if err != nil {
		log.Fatalf("Unable to retrieve token from web: %v", err)
	}
	return tok
}

func tokenFromFile(file string) (*oauth2.Token, error) {
	f, err := os.Open(file)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	tok := &oauth2.Token{}
	err = json.NewDecoder(f).Decode(tok)
	return tok, err
}

func saveToken(path string, token *oauth2.Token) {
	fmt.Printf("Saving credential file to: %s\n", path)
	f, err := os.OpenFile(path, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		log.Fatalf("Unable to cache oauth token: %v", err)
	}
	defer f.Close()
	json.NewEncoder(f).Encode(token)
}
