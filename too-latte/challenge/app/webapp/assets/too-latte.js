document.addEventListener('DOMContentLoaded',() => {
  document.getElementById("need-token").addEventListener("click", () => {
    var xhttp = new XMLHttpRequest();

    xhttp.onload = () => {
      document.getElementById("generated-request").innerHTML = xhttp.responseText;
      document.getElementById("need-token-response").style.display = "block";
      document.getElementById("have-token-response").style.display = "none";
    };
    xhttp.open("GET", "/generate", true);
    xhttp.send();
  });

  document.getElementById("have-token").addEventListener("click", () => {
    document.getElementById("need-token-response").style.display = "none";
    document.getElementById("have-token-response").style.display = "block";
  });

  document.getElementById("submit-token").addEventListener("click", () => {
    var xhttp = new XMLHttpRequest();

    xhttp.onload = () => {
      document.getElementById("response").innerHTML = xhttp.responseText;
    };
    xhttp.open("GET", "/validate?token=" + document.getElementById("entered-token").value, true);
    xhttp.send();
  });
});
