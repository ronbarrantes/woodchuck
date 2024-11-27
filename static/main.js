const logUL = document.querySelector(".log-list");
const exportBtn = document.querySelector(".export-button");

exportBtn.addEventListener("click", () => {
  console.log("CLICKED AND EXPORTINT, NOT REALLY LOL");
});

logUL.style.background = "green";
const url = "http://localhost:8080/api/v1/logs";

const colorLevel = (level) => {
  let levelColor = "fff";
  switch (level) {
    case "warn":
      levelColor = "#ff0";
      break;
    case "error":
      levelColor = "#f00";
      break;
    default:
      break;
  }
  return levelColor;
};

const createLi = (timestamp, log_id, level, user_id, message) => {
  const logLI = document.createElement("li");
  const tsLi = document.createElement("span");
  const lLvlLi = document.createElement("span");
  const lIdLi = document.createElement("span");
  const uIdLi = document.createElement("span");
  const msgLi = document.createElement("span");

  logLI.className = "log-list-li";
  lLvlLi.style.color = colorLevel(level);

  tsLi.innerText = timestamp;
  lLvlLi.innerText = level;
  lIdLi.innerText = log_id;
  uIdLi.innerText = user_id;
  msgLi.innerText = message;

  logLI.appendChild(tsLi);
  logLI.appendChild(lLvlLi);
  logLI.appendChild(lIdLi);
  logLI.appendChild(uIdLi);
  logLI.appendChild(msgLi);

  return logLI;
};

// /api/v1/logs
const getLogs = async () => {
  const apiCall = await fetch(url);
  data = (await apiCall.json()).map((log) => {
    const date = new Date(log.timestamp);
    console.log("TS", log.timestamp);
    timestamp = date.toISOString();

    const logLi = createLi(
      timestamp,
      log.log_id,
      log.level,
      log.user_id,
      log.message,
    );

    logUL.appendChild(logLi);
  });
};

const eventSource = new EventSource("http://localhost:8080/api/v1/events");

eventSource.onmessage = (event) => {
  console.log("New SSE message:", event.data);
};

getLogs();
