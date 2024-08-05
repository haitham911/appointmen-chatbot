
# Appointment Booking Bot GPT-4

This project implements an appointment booking assistant using WebSockets and OpenAI's GPT-4 model. The bot can manage conversations to book appointments, handle user messages, and keep track of session states.

## Features

- WebSocket-based real-time communication
- State management for user sessions
- Integration with OpenAI's GPT-4 for natural language understanding
- Dynamic response generation based on user inputs
- Mock data for departments, doctors, and available times

## Project Structure

```
.
├── main.go              # Main application entry point
├── .env                 # Environment variables file
├── go.mod               # Go module file
└── go.sum               # Go dependencies file
```

## Getting Started

### Prerequisites

- Go 1.22.5+
- Docker (for running in a containerized environment, optional)

### Installation

1. Clone the repository:

   ```sh
   git clone https://github.com/your-username/appointment-booking-bot.git
   cd appointment-booking-bot
   ```

2. Create a `.env` file in the project root directory with your OpenAI API key:

   ```sh
   OPENAI_API_KEY=your-openai-api-key
   ```

3. Install dependencies:

   ```sh
   go mod tidy
   ```

### Running the Application

1. Start the application:

   ```sh
   go run main.go
   ```

2. The server will start on port `8080`. Connect to it using a WebSocket client at `ws://localhost:8080/ws`.

### Environment Variables

- `OPENAI_API_KEY`: Your OpenAI API key.

### Usage

The bot manages user sessions to book appointments with the following states:

1. `start`: Initial state.
2. `awaiting_client_status`: Asking if the user is a new or existing client.
3. `awaiting_email_new`: Asking for the email address of a new client.
4. `awaiting_email_existing`: Asking for the email address of an existing client.
5. `awaiting_name`: Asking for the name of a new client.
6. `awaiting_department`: Asking for the department for the appointment.
7. `fetch_doctor_department`: Fetching available doctors for the selected department.
8. `awaiting_doctor`: Asking for the preferred doctor.
9. `awaiting_time`: Asking for the preferred appointment time.
10. `awaiting_address_option`: Asking if the user wants to provide a pickup address.
11. `awaiting_address`: Asking for the pickup address.
12. `completed`: Final state when the appointment is successfully booked.

### Example Conversation

```
User: start
Bot: Hello, are you a new client or existing client? Please respond with 'new' or 'existing'.

User: new
Bot: Please provide your email address.

User: user@example.com
Bot: Please provide your name.

User: John Doe
Bot: Which department would you like to book an appointment with? Available departments are: A - Cardio, B - Ortho, C - Neuro, D - Gync, E - General.

User: B
Bot: Available doctors in Ortho department: A - Dr. Jagadeesh MS, B - Dr. Ravikiran MS, C - Dr. Haitham MS. Please select a doctor.

User: A
Bot: Available times for appointments are: A - 09:00, B - 10:00, C - 11:00, D - 12:00. Please select a time.

User: B
Bot: Would you like to provide an address for pickup? (yes/no)

User: yes
Bot: Please provide your address.

User: 123 Main St
Bot: Your appointment has been booked successfully!
```

### Dependencies

- `github.com/gorilla/websocket`: WebSocket library for Go.
- `github.com/joho/godotenv`: Library to load environment variables from a `.env` file.
- `github.com/sashabaranov/go-openai`: Go client for OpenAI API.

### License

This project is licensed under the MIT License. See the [LICENSE](LICENSE) file for details.

### Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

### Acknowledgements

- OpenAI for their powerful GPT-4 model.
- Gorilla WebSocket library for handling WebSocket connections.

## Contact

For any questions or suggestions, please reach out to [your-email@example.com](mailto:your-email@example.com).
