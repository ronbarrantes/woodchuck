const h1El = document.getElementsByTagName("h1")[0];

h1El.style.background = "blue";
console.log("from main.js");

const getLogs = async () => {
  const apiCall = await fetch("/api/v1/logs");
  data = (await apiCall.json()).map((log) => {
    const date = new Date(log.timestamp);
    timestamp = date.toISOString();
    return { ...log, timestamp };
  });

  console.log(data);
};

// getLogs();
