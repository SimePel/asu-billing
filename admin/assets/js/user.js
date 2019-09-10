import goBack from "./goBack.js";

const urlParams = new URLSearchParams(window.location.search);
const userID = urlParams.get("id");
getUser(userID);

let paymentButton = document.querySelector("#paymentButton");
paymentButton.addEventListener("click", revealPaymentInput);

let deleteButton = document.querySelector("#deleteButton");
deleteButton.addEventListener("click", deleteUser);

function getUser(userID) {
    fetch("/users/" + userID).then((res) => { return res.json() }).then((user) => {
        document.querySelector("#name").append(user.name);
        document.querySelector("#agreement").append(user.agreement);
        document.querySelector("#login").append(user.login);
        document.querySelector("#tariff").append(user.tariff.name);
        document.querySelector("#innerIP").append(user.inner_ip);
        document.querySelector("#extIP").append(user.ext_ip);
        document.querySelector("#phone").append(user.phone);
        document.querySelector("#room").append(user.room);
        document.querySelector("#connectionPlace").append(user.connection_place);
        if (user.activity === true) {
            const d = new Date(user.expired_date);
            const expiredDate = d.getDate() + "." + (d.getMonth() + 1) + "." + d.getFullYear();
            document.querySelector("#expiredDate").append(expiredDate);
        } else {
            document.querySelector("#expiredDate").parentElement.remove();
        }
        document.querySelector("#balance").append(user.balance);
    })
}

function deposit() {
    fetch("payment", {
        method: "POST",
        headers: { "Content-Type": "application/json; charset=utf-8" },
        body: JSON.stringify({
            id: parseInt(userID),
            sum: parseInt(document.querySelector("#paymentInput").value),
        }),
    }).then(() => {
        goBack();
    });
}

function revealPaymentInput() {
    let paymentInput = document.querySelector("#paymentInput");
    let paymentInputParent = paymentInput.parentElement;

    paymentInputParent.removeAttribute("hidden");
    paymentButton.removeEventListener("click", revealPaymentInput);

    paymentButton.addEventListener("click", deposit);
    paymentInput.addEventListener("keyup", event => {
        if (event.keyCode === 13) {
            deposit();
        }
    });
}

function deleteUser() {
    let answer = confirm("Вы действительно хотите удалить этого пользователя?");
    if (answer === true) {
        fetch("users/" + userID, {
            method: "DELETE",
        }).then(() => {
            goBack();
        });
    }
}
