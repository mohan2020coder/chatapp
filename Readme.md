# **Real-Time Chat Application - Project Documentation**

### **Overview**

This project is a **Real-Time Chat Application** that allows users to send and receive messages instantly using WebSockets. It includes user authentication with JWT tokens, support for both group chat and private messaging, and the persistence of messages in a database.

### **Technologies Used**

- **Backend**: Go (Golang)
- **WebSockets**: For real-time messaging
- **Authentication**: JWT (JSON Web Tokens)
- **Database**: PostgreSQL (or any relational DB for message storage)
- **Frontend**: HTML, CSS, JavaScript (React for frontend or vanilla JS for simplicity)
- **Libraries**:
  - `gorilla/websocket` for WebSocket handling in Go
  - `jwt-go` for JWT token creation/verification
  - `Gorm` for ORM and database interaction

---

## **System Architecture**

### **High-Level Components**

1. **Client**:
   - Web Browser (HTML/JS or React)
   - Can connect via WebSocket to the server
   - Displays messages in real time
   - Allows user to send messages (both private and group)

2. **Server (Go Backend)**:
   - Handles WebSocket connections
   - Manages message broadcasting
   - Manages JWT-based user authentication
   - Persists messages in the database
   - Handles private and group chats

3. **Database**:
   - Stores user information
   - Stores message history
   - Tables:
     - **Users**: Stores user information (ID, username, password hash, etc.)
     - **Messages**: Stores chat messages (message ID, sender ID, receiver ID, content, timestamp)
     - **Groups**: Stores group chat details (group ID, group name, member list)

---

## **Functional Requirements**

1. **User Registration**:
   - A user should be able to register with a username and password.
   - Passwords should be hashed before being stored in the database.
   
2. **User Authentication**:
   - Users should authenticate via JWT tokens. Tokens should expire after a set period of time and can be refreshed.
   - Users must send the JWT in the authorization header when making API requests.
   
3. **WebSocket Connection**:
   - Users connect to the WebSocket server upon successful login.
   - The server listens for incoming messages from clients and broadcasts to relevant recipients (either a group or a private message).

4. **Group and Private Chat**:
   - Users should be able to join group chats or send private messages to other users.
   - Messages should be sent and received in real time.

5. **Message History**:
   - All messages should be stored in the database.
   - Users can retrieve message history when they connect, for both group chats and private chats.

---

## **System Design**

### **1. WebSocket Server (Go)**

1. **WebSocket Connection Handler**: 
   - The server listens for incoming WebSocket connections.
   - Once a client connects, the server authenticates the user based on the JWT token sent in the WebSocket request header.
   - After successful authentication, the server assigns the client a WebSocket connection ID.
   
2. **Message Broadcasting**:
   - The server listens for incoming messages and broadcasts them to either a specific user (private chat) or a group of users (group chat).

3. **User Management**:
   - Track online users and their active WebSocket connections.
   - Provide mechanisms to send direct messages (private messages) or broadcast to a group chat.

4. **Message Persistence**:
   - After receiving a message, the server saves it to the database with relevant metadata (sender, receiver, content, timestamp).

---

### **2. Database Schema**

- **Users Table**:
```sql
  CREATE TABLE users (
      id SERIAL PRIMARY KEY,
      username VARCHAR(255) UNIQUE NOT NULL,
      password_hash VARCHAR(255) NOT NULL,
      created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
  )
  ```
- **Messages Table**:
```sql
CREATE TABLE messages (
    id SERIAL PRIMARY KEY,
    sender_id INT REFERENCES users(id),
    receiver_id INT,
    group_id INT,
    message TEXT NOT NULL,
    timestamp TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);
```
- **Groups Table**:
```sql
CREATE TABLE groups (
    id SERIAL PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);
```
