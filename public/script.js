document.addEventListener("DOMContentLoaded", () => {
    const canvas = document.getElementById("drawingCanvas");
    const ctx = canvas.getContext("2d");
    let drawing = false;
    let lastX, lastY;

    const clientID = Math.random().toString(36).substr(2, 9);
    const socketProtocol = location.protocol.includes("https") ? "wss" : "ws";
    const ws = new WebSocket(`${socketProtocol}://${window.location.host}${window.location.pathname}ws`);

    ws.onmessage = function(event) {
        const data = JSON.parse(event.data);
        if (data.clientID !== clientID) {
            drawLine(data.x0 * canvas.width, data.y0 * canvas.height, data.x1 * canvas.width, data.y1 * canvas.height, false);
        }
    };

    canvas.addEventListener("mousedown", (e) => {
        drawing = true;
        [lastX, lastY] = [e.offsetX, e.offsetY];
    });

    canvas.addEventListener("mouseup", () => {
        drawing = false;
    });

    canvas.addEventListener("mousemove", (e) => {
        if (!drawing) return;
        const [x, y] = [e.offsetX, e.offsetY];
        drawLine(lastX, lastY, x, y, true);
        [lastX, lastY] = [x, y];
    });

    function drawLine(x0, y0, x1, y1, emit) {
        ctx.beginPath();
        ctx.moveTo(x0, y0);
        ctx.lineTo(x1, y1);
        ctx.strokeStyle = "black";
        ctx.lineWidth = 2;
        ctx.stroke();
        ctx.closePath();

        if (!emit) return;
        const w = canvas.width;
        const h = canvas.height;

        ws.send(JSON.stringify({
            clientID: clientID,
            x0: x0 / w,
            y0: y0 / h,
            x1: x1 / w,
            y1: y1 / h
        }));
    }
});

