var webSocket = null;
let username = "";
var timerStarted = false;

window.onload = function () {
  if (!timerStarted) {

    const app = document.getElementById("app");
    const usernameContainer = createElement("div", {
      id: "username-container",
    });
    const usernameInput = createElement("input", {
      type: "text",
      id: "username-input",
      placeholder: "Enter your username",
    });
    const submitButton = createElement("button", { id: "button" }, ["Submit"]);

    appendChild(usernameContainer, usernameInput);
    appendChild(usernameContainer, submitButton);
    appendChild(app, usernameContainer);

    onClick(submitButton, function () {
      username = usernameInput.value.trim();
      if (username !== "") {
        initWebSocket();
      } else {
        alert("Please enter a valid username.");
      }
    });

    // this is a eventlistener for handling movements of the player
    document.addEventListener("keydown", handleKeyPress);
  } else {
    alert("you can't join the game now, it has already started");
  }
};

function initWebSocket() {
  const socket = new WebSocket(`ws://localhost:5050/ws?username=${username}`);
  webSocket = socket;
  webSocket.onopen = function (event) {
    console.log("WebSocket is open now.");
    webSocket.send(
      JSON.stringify({
        type: "add_socket",
        sender: username,
      })
    );
  };

  webSocket.onmessage = function (event) {
    const socketData = JSON.parse(event.data);
    const { type } = socketData;

    switch (type) {
      case "user_added":
        createGameUI();
        break;
      case "user_taken":
        alert("Username is already taken. Please choose another.");
        break;
      case "server_full":
        alert("Server is full, try again later.");
        break;
      case "player_list":
        addPlayersToLobby(socketData.usernames);
        break;
      case "chat":
        displayMessage(socketData.sender, socketData.message);
        break;
      case "timer":
        displayCounter(socketData.timer);
        break;
      case "loadBoard":
        createBoard();
        updateGrid(socketData.board);
        updatePlayerPosition(socketData.players);
        break;
      case "playerAction":
        updatePlayerPosition(socketData.players);
        updateGrid(socketData.board);
        break;
      case "bomb":
      case "explosion":
        updateGrid(socketData.board);
        break;
    }
  };

  // Handling connection errors
  webSocket.onerror = function (event) {
    console.error("WebSocket error:", event);
  };

  // Connection closed
  webSocket.onclose = function (event) {
    console.log("WebSocket is closed now.");
  };

  return socket;
}

// Function to update the grid based on game state
function updateGrid(board) {
  const gameBoard = document.getElementById("game-board");
  // Update grid based on game state
  for (let row = 0; row < 13; row++) {
    for (let col = 0; col < 15; col++) {
      const cell = gameBoard.querySelector(
        `[data-row="${row}"][data-col="${col}"]`
      );

      switch (board[row][col]) {
        case " ":
          cell.classList.remove(
            "wall",
            "block",
            "bomb",
            "blast-range",
            "power-up-bombs",
            "power-up-flames",
            "power-up-speed"
          );
          break;
        case "X":
          cell.classList.add("wall");
          break;
        case "B":
          cell.classList.add("block");
          break;
        case "bomb":
          cell.classList.add("bomb");
          break;
        case "#":
          cell.classList.add("blast-range");
          break;
        case "power-up-bombs":
          cell.classList.remove("bomb");
          cell.classList.add("power-up-bombs");
          break;
        case "power-up-flames":
          cell.classList.remove("bomb");
          cell.classList.add("power-up-flames");
          break;
        case "power-up-speed":
          cell.classList.remove("bomb");
          cell.classList.add("power-up-speed");
          break;
      }
    }
  }
}

function createGameUI() {
  const app = document.getElementById("app");
  removeAllChildren(app);
  const gameContainer = createElement("div", { id: "game-container" });
  const gameBoard = createElement("div", { id: "game-board" });
  const lobbyContainer = createElement("div", { id: "lobby-container" });
  const listHeader = createElement("div", { id: "list-header" }, ["Players:"]);
  const playerList = createElement("div", { id: "player-list" }, []);
  appendChild(app, gameContainer);
  appendChild(gameContainer, gameBoard);
  appendChild(lobbyContainer, listHeader);
  appendChild(lobbyContainer, playerList);

  const chatContainer = createElement("div", { id: "chat-container" });
  const messageDisplay = createElement("div", { id: "message-display" });
  appendChild(chatContainer, messageDisplay);

  const messageInput = createElement("input", {
    id: "message-input",
    type: "text",
    placeholder: "Type your message...",
  });
  appendChild(chatContainer, messageInput);

  const sendButton = createElement("button", { id: "button" }, ["Send"]);
  appendChild(chatContainer, sendButton);

  onClick(sendButton, function () {
    sendMessage("chat", messageInput.value);
    messageInput.value = "";
  });

  appendChild(app, chatContainer);
  appendChild(chatContainer, lobbyContainer);

  // Event listener for the Enter key in the chat input
  document
    .getElementById("message-input")
    .addEventListener("keypress", function (e) {
      // Check if Enter key is pressed
      if (e.key === "Enter") {
        // Send chat message to the backend via WebSocket
        sendMessage("chat", this.value);
        this.value = ""; // Clear the chat input after sending the message
      }
    });
}

function addPlayersToLobby(usernames) {
  const playerList = document.getElementById("player-list");
  removeAllChildren(playerList);

  usernames.forEach((username) => {
    const player = createElement("div", { class: "player" }, [username]);
    appendChild(playerList, player);
  });
}

function sendMessage(type, message) {
  webSocket.send(
    JSON.stringify({
      type: type,
      sender: username,
      message: message,
    })
  );
}

function displayMessage(sender, message) {
  const messageDisplay = document.getElementById("message-display");
  const messageElement = createElement("div", {}, [`${sender}: ${message}`]);
  appendChild(messageDisplay, messageElement);
}

function displayCounter(counter) {
  const gameBoard = document.getElementById("game-board");
  if (!gameBoard) {
    console.error('Element with ID "game-board" not found');
    return;
  }
  timerStarted = true;

  removeAllChildren(gameBoard);
  const counterDiv = createElement("div", { id: "countdown" }, [`${counter}`]);
  appendChild(gameBoard, counterDiv);
}

function createBoard() {
  const gameBoard = document.getElementById("game-board");
  removeAllChildren(gameBoard); // Clear previous content

  for (let row = 0; row < 13; row++) {
    for (let col = 0; col < 15; col++) {
      gameBoard.appendChild(
        createElement("div", {
          class: "grid-cell",
          "data-row": row,
          "data-col": col,
        })
      );
    }
  }
}

function updatePlayerPosition(players) {
  const gameBoard = document.getElementById("game-board");

  // Remove all classes that start with "player" from all grid cells
  gameBoard.querySelectorAll("[class*='player']").forEach((playerCell) => {
    const classNames = playerCell.className.split(" ");
    const filteredClasses = classNames.filter(
      (cls) => !cls.startsWith("player")
    );
    playerCell.className = filteredClasses.join(" ");
  });

  // Update player positions
  players.forEach((player) => {
    const playerCell = gameBoard.querySelector(
      `[data-row="${player.position.x}"][data-col="${player.position.y}"]`
    );
    if (playerCell) {
      playerCell.classList.add(`player${player.id}`);
    }
  });
}

// Function to handle key press events and send directly to server
function handleKeyPress(e) {
  // Check if the focus is on the chat input
  const isChatInputFocused = document.activeElement.id === "message-input";
  // If the focus is on the chat input, do not send key press to server
  if (isChatInputFocused) {
    return;
  }

  // Check if the key pressed is an arrow key or spacebar
  if (
    e.key === "ArrowLeft" ||
    e.key === "ArrowUp" ||
    e.key === "ArrowRight" ||
    e.key === "ArrowDown" ||
    e.key === " "
  ) {
    // Send key press information to the backend via WebSocket
    webSocket.send(
      JSON.stringify({
        type: "keypress",
        key: e.key,
        user_name: username,
      })
    );
  }
}
