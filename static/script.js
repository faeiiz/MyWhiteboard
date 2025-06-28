// Prompt for name
let username = "";
while (!username) {
  username = prompt("Enter your name");
}

const allStrokes = [];

// Canvas setup (unchanged)
const canvas = document.getElementById("board");
const ctx = canvas.getContext("2d");
canvas.width = window.innerWidth;
canvas.height = window.innerHeight;
window.addEventListener("resize", () => {
  canvas.width = window.innerWidth;
  canvas.height = window.innerHeight;
});

// WebSocket
const socket = new WebSocket("ws://" + window.location.host + "/ws");
let userId = null; // will be populated by server

socket.onopen = () => {
  console.log("Connected to WebSocket");
  // Send join message immediately
  socket.send(
    JSON.stringify({
      type: "join",
      username: username,
    })
  );
};

socket.onmessage = (event) => {
  const data = JSON.parse(event.data);
 console.log("Received message:", data); 
  switch (data.type) {
    case "joined":
      // Server confirms your ID
      userId = data.userId;
      console.log("Assigned userId:", userId);
      break;

    case "draw":
      allStrokes.push(data);
      drawLine(
        data.fromX,
        data.fromY,
        data.toX,
        data.toY,
        data.color,
        data.size
      );
      break;

    case "clear":
      clearCanvas();
      break;

    case "userCount":
      updateUserCount(data.count);
      break;

    default:
      console.warn("Unknown message type:", data.type);
  }
};

socket.onclose = () => {
  console.log("Disconnected from WebSocket");
};

// Controls
const colorPicker = document.getElementById("colorPicker");
const sizePicker = document.getElementById("sizePicker");
const clearBtn = document.getElementById("clearBtn");

let drawing = false;
let lastX = 0;
let lastY = 0;

// Drawing listeners
canvas.addEventListener("mousedown", (e) => {
  drawing = true;
  lastX = e.clientX;
  lastY = e.clientY;
});
canvas.addEventListener("mouseup", () => (drawing = false));
canvas.addEventListener("mousemove", (e) => {
  if (!drawing || !userId) return; // wait for join
  const currX = e.clientX,
    currY = e.clientY;
  const color = colorPicker.value;
  const size = parseInt(sizePicker.value, 10);

  // Local draw
  drawLine(lastX, lastY, currX, currY, color, size);

  // Broadcast with your userId & username
  socket.send(
    JSON.stringify({
      type: "draw",
      userId: userId,
      username: username,
      fromX: lastX,
      fromY: lastY,
      toX: currX,
      toY: currY,
      color: color,
      size: size,
    })
  );

  lastX = currX;
  lastY = currY;
});

// Clear button
clearBtn.addEventListener("click", () => {
  if (!userId) return;
  clearCanvas();
  socket.send(
    JSON.stringify({
      type: "clear",
      userId: userId,
      username: username,
    })
  );
});

// Utility functions
function drawLine(x1, y1, x2, y2, color, size) {
  ctx.strokeStyle = color;
  ctx.lineWidth = size;
  ctx.lineCap = "round";
  ctx.beginPath();
  ctx.moveTo(x1, y1);
  ctx.lineTo(x2, y2);
  ctx.stroke();
}

function clearCanvas() {
  ctx.clearRect(0, 0, canvas.width, canvas.height);
}

function updateUserCount(count) {
  const span = document.getElementById("userCount");
  span.textContent = `${count} user${count === 1 ? "" : "s"} online`;
}
