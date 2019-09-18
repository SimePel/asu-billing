const urlParams = new URLSearchParams(window.location.search);
const userID = urlParams.get("id");

function insertValuesInInputs() {
    fetch("/users/" + userID).then((res) => {
        return res.json()
    }).then((user) => {
        document.getElementsByName("id")[0].value = user.id;
        document.getElementsByName("name")[0].value = user.name;
        document.getElementsByName("agreement")[0].value = user.agreement;
        document.getElementsByName("login")[0].value = user.login;
        document.getElementsByName("phone")[0].value = user.phone;
        document.getElementsByName("room")[0].value = user.room;
        document.getElementsByName("connectionPlace")[0].value = user.connection_place;
    })
}

insertValuesInInputs();

let cancelButton = document.querySelector("#cancelButton");
cancelButton.setAttribute("href", "/user?id=" + userID);