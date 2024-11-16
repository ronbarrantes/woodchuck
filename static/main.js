const h1El = document.getElementsByTagName("h1")[0];

h1El.style.background = "blue";
console.log("from main.js");

const getLogs = async () => {
  const apiCall = await fetch("/api/v1/logs");
  const data = await apiCall.text();
  console.log("TEXT--->", data);
};

getLogs();
