const canvas = document.getElementById("board");
const ctx = canvas.getContext("2d");

canvas.width = window.innerWidth;
canvas.height = window.innerHeight;

// Adjust canvas on resize
window.addEventListener("resize", () => {
  canvas.width = window.innerWidth;
  canvas.height = window.innerHeight;
});

// WebSocket connection
const socket = new WebSocket("ws://" + window.location.host + "/ws");

socket.onopen = () => {
  console.log("Connected to WebSocket");
};

socket.onmessage = (event) => {
  const data = JSON.parse(event.data);
  if (data.type === "draw") {
    drawCircle(data.x, data.y, data.color, data.size);
  }
};

socket.onclose = () => {
  console.log("Disconnected from WebSocket");
};

// Drawing settings
const color = "#000000";
const size = 2;

let drawing = false;

canvas.addEventListener("mousedown", (e) => {
  drawing = true;
  drawAndSend(e);
});

canvas.addEventListener("mouseup", () => {
  drawing = false;
});

canvas.addEventListener("mousemove", (e) => {
  if (drawing) {
    drawAndSend(e);
  }
});

function drawAndSend(e) {
  const x = e.clientX;
  const y = e.clientY;

  // Draw locally
  drawCircle(x, y, color, size);

  // Send to server
  const msg = {
    type: "draw",
    x: x,
    y: y,
    color: color,
    size: size
  };
  socket.send(JSON.stringify(msg));
}

function drawCircle(x, y, color, size) {
  ctx.fillStyle = color;
  ctx.beginPath();
  ctx.arc(x, y, size, 0, Math.PI * 2);
  ctx.fill();
}
