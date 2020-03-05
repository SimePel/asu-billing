const urlParams = new URLSearchParams(window.location.search);
const userID = urlParams.get("id");

function insertValuesInInputs() {
  fetch("/users/" + userID)
    .then(res => {
      return res.json();
    })
    .then(user => {
      document.getElementsByName("id")[0].value = user.id;
      document.getElementsByName("name")[0].value = user.name;
      document.getElementsByName("agreement")[0].value = user.agreement;

      if (user.is_employee === true) {
        document.getElementsByName("isEmployee")[0].checked = false;
        document.getElementsByName("isEmployee")[1].checked = true;
      }

      document.getElementsByName("login")[0].value = user.login;
      document.getElementsByName("mac")[0].value = user.mac;
      document.getElementsByName("phone")[0].value = user.phone;
      document.getElementsByName("room")[0].value = user.room;
      document.getElementsByName("connectionPlace")[0].value = user.connection_place;
      document.getElementsByName("comment")[0].value = user.comment;
      const d = new Date(user.expired_date);
      let day = d.getDate();
      if (day < 10) day = "0" + day;
      let month = d.getMonth() + 1;
      if (month < 10) month = "0" + month;
      const year = d.getFullYear();
      let expiredDate = year + "-" + month + "-" + day;
      document.getElementsByName("expiredDate")[0].value = expiredDate;
    });
}

insertValuesInInputs();

let cancelButton = document.querySelector("#cancelButton");
cancelButton.setAttribute("href", "/user?id=" + userID);
