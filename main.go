package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"sync"

	"github.com/gorilla/websocket"
	"github.com/joho/godotenv"
	openai "github.com/sashabaranov/go-openai"
)

type Session struct {
	State      string
	Email      string
	Name       string
	Department string
	Doctor     string
	Time       string
	Address    string
	Messages   []openai.ChatCompletionMessage
}

var (
	sessions     = make(map[string]*Session)
	sessionsLock sync.Mutex
	upgrader     = websocket.Upgrader{
		CheckOrigin: func(r *http.Request) bool {
			return true
		},
	}
	client *openai.Client
)

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	client = openai.NewClient(os.Getenv("OPENAI_API_KEY"))
	http.HandleFunc("/ws", handleConnections)
	fmt.Println("Server started on :8080")
	err = http.ListenAndServe(":8080", nil)
	if err != nil {
		fmt.Println("ListenAndServe error:", err)
	}
}

func handleConnections(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		fmt.Println("Upgrade error:", err)
		return
	}
	defer conn.Close()

	userID := conn.RemoteAddr().String()
	log.Println("Connected:", userID)

	for {
		_, message, err := conn.ReadMessage()
		if err != nil {
			log.Println("ReadMessage error:", err)
			break
		}

		response, err := handleUserMessage(client, userID, string(message))
		if err != nil {
			log.Println("handleUserMessage error:", err)
			break
		}

		err = conn.WriteMessage(websocket.TextMessage, []byte(response))
		if err != nil {
			log.Println("WriteMessage error:", err)
			break
		}
	}
	log.Println("Connection closed for user:", userID)
}

func handleUserMessage(client *openai.Client, userID string, userMessage string) (string, error) {
	sessionsLock.Lock()
	session, exists := sessions[userID]
	if !exists {
		session = &Session{State: "start"}
		sessions[userID] = session
	}
	sessionsLock.Unlock()

	prompt := fmt.Sprintf("Current state: %s\nUser message: %s\n\nDetermine the next state and response based on the current state and user message.\n\n", session.State, userMessage)

	response, err := callOpenAI(client, prompt, "", session)
	if err != nil {
		return "", err
	}
	fmt.Println(response)
	nextState, nextResponse := parseOpenAIResponse(response)
	if nextState == "fetch_doctor_department" {
		sessionsLock.Lock()
		session.State = nextState
		sessionsLock.Unlock()
		department := strings.TrimPrefix(nextResponse, "return selected department ")
		department = strings.ReplaceAll(department, `"`, "")
		doctors := getDoctorsByDepartment(department)
		prompt = fmt.Sprintf("Current state: %s\nUser message: %s\n\nDetermine the next state and response based on the current state and user message.\n\n", "fetch_doctor_department", "what is current  doctor available on department "+department)

		d := strings.Join(doctors, ",")
		response, err = callOpenAI(client, prompt, d, session)
		if err != nil {
			return "", err
		}
		nextState, nextResponse = parseOpenAIResponse(response)
		fmt.Println(response)

	}
	sessionsLock.Lock()
	session.State = nextState
	sessionsLock.Unlock()

	return nextResponse, nil
}

func callOpenAI(client *openai.Client, message string, messageBackend string, session *Session) (string, error) {
	ctx := context.Background()
	systemMessage := getExample()
	// Inject doctors and times dynamically based on the state
	if session.State == "fetch_doctor_department" && messageBackend != "" {
		fmt.Println(messageBackend)
		systemMessage += fmt.Sprintf("\n - use current department Available doctors : %s", messageBackend)
	} else if session.State == "awaiting_doctor" && messageBackend != "" {
		times := getAvailableTimesForDoctor(session.Doctor)
		systemMessage += fmt.Sprintf("\nAvailable times: %s", strings.Join(times, ", "))
	}
	req := openai.ChatCompletionRequest{
		Model: "gpt-4",
		Messages: []openai.ChatCompletionMessage{
			{
				Role:    openai.ChatMessageRoleSystem,
				Content: systemMessage,
			},
			{
				Role:    openai.ChatMessageRoleUser,
				Content: message,
			},
		},
	}
	resp, err := client.CreateChatCompletion(ctx, req)
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(resp.Choices[0].Message.Content), nil
}

func parseOpenAIResponse(response string) (string, string) {
	lines := strings.Split(response, "\n")
	nextState := strings.TrimPrefix(lines[0], "Next state: ")
	nextResponse := strings.TrimPrefix(lines[1], "Response: ")
	return strings.TrimSpace(nextState), strings.TrimSpace(nextResponse)
}

// Dummy implementation for checking if client exists
func clientExists(email string) bool {
	return email == "existing@example.com"
}

// Dummy implementation for retrieving doctors by department
func getDoctorsByDepartment(department string) []string {
	if department == "Ortho" {
		return []string{"Dr. Jagadeesh MS", "Dr. Ravikiran MS", "Dr. Haitham MS"}
	}
	return []string{"Dr. Smith", "Dr. Jones"}
}

// Dummy implementation for retrieving available times for a doctor
func getAvailableTimesForDoctor(doctor string) []string {
	return []string{"09:00", "10:00", "11:00", "12:00"}
}

// Dummy implementation for booking an appointment
func bookAppointment(session *Session) (string, error) {
	return "Appointment booked successfully with " + session.Doctor + " at " + session.Time, nil
}

func getExample() string {
	examples := `
You are an appointment booking assistant.
You must return the next state and response.
Move the user between states as necessary.
on error didn't understand repeat last state

State: start
User message: start or any word 
Next state: awaiting_client_status
Response: Hello , "Are you a new client or existing client? Please respond with 'new' or 'existing'."

---

State: awaiting_client_status
User message: new
Next state: awaiting_email_new
Response: "Please provide your email address."

---

State: awaiting_client_status
User message: existing
Next state: awaiting_email_existing
Response: "Please provide your email address."

---

State: awaiting_email_new
User message: user@example.com
Next state: awaiting_name
Response: "Please provide your name."

---

State: awaiting_email_existing
User message: user@example.com
Next state: awaiting_department
Response: "Welcome back! Which department would you like to book an appointment with? Available departments are: A - Cardio, B - Ortho, C - Neuro, D - Gync, E - General."

---

State: awaiting_name
User message: John Doe
Next state: awaiting_department
Response: "Which department would you like to book an appointment with? Available departments are: A - Cardio, B - Ortho, C - Neuro, D - Gync, E - General."

---

State: awaiting_department
User message: A
Next state: fetch_doctor_department
Response: return selected department "Cardio"

---
State: fetch_doctor_department
User message: 
Next state: awaiting_doctor
Response: "Available doctors in Cardio department: A - Dr. Smith, B - Dr. Jones. Please select a doctor."

---
State: awaiting_doctor
User message: A
Next state: awaiting_time
Response: "Available times for appointments are: A - 09:00, B - 10:00, C - 11:00, D - 12:00. Please select a time."

---

State: awaiting_time
User message: A
Next state: awaiting_address_option
Response: "Would you like to provide an address for pickup? (yes/no)"

---

State: awaiting_address_option
User message: yes
Next state: awaiting_address
Response: "Please provide your address."

---

State: awaiting_address_option
User message: no
Next state: completed
Response: "Your appointment has been booked successfully!"

---

State: awaiting_address
User message: 123 Main St
Next state: completed
Response: "Your appointment has been booked successfully!"

---

The departments include:
- Cardio
- Ortho
- Neuro
- Gync
- General

The doctors include:
- Dr. Jagadeesh MS, Ortho
- Dr. Ravikiran MS, Ortho
- Dr. Jaggu MS, Cardio
- Dr. Sanju MS, Gync
- Dr. Tanishq MS, General

The timings include:
- 09:00
- 10:00
- 11:00
- 12:00

`
	return examples
}
