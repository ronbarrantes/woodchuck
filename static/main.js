const logUl = document.querySelector(".log-list");
const exportBtn = document.querySelector(".export-button");

exportBtn.addEventListener("click", () => {
  console.log("CLICKED AND EXPORTINT, NOT REALLY LOL");
});

// logUl.style.background = "green";
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

const createLi = (log) => {
  console.log(log);

  const logLi = document.createElement("li");
  const tsLi = document.createElement("span");
  const lLvlLi = document.createElement("span");
  const lIdLi = document.createElement("span");
  const uIdLi = document.createElement("span");
  const msgLi = document.createElement("span");

  const date = new Date(log.timestamp);
  const ts = date.toISOString();

  logLi.className = "log-item";
  lLvlLi.style.color = colorLevel(log.level);

  tsLi.innerText = ts;
  lLvlLi.innerText = log.level;
  lIdLi.innerText = log.log_id;
  uIdLi.innerText = log.user_id;
  msgLi.innerText = log.message;

  logLi.appendChild(tsLi);
  logLi.appendChild(lLvlLi);
  logLi.appendChild(lIdLi);
  logLi.appendChild(uIdLi);
  logLi.appendChild(msgLi);

  return logLi;
};

// /api/v1/logs
const getLogs = async () => {
  const apiCall = await fetch(url);
  data = (await apiCall.json()).map((log) => {
    const logLi = createLi(log);

    logUl.appendChild(logLi);
  });

  logUl.lastChild.scrollIntoView({ behavior: "instant" });
};

const eventSource = new EventSource("http://localhost:8080/api/v1/events");

eventSource.onmessage = (event) => {
  console.log(typeof event.data);
  const logLi = createLi(JSON.parse(event.data));
  logUl.append(logLi);
  logLi.scrollIntoView({ behavior: "smooth" });
};

getLogs();
