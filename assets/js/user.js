const urlParams = new URLSearchParams(window.location.search);
const userID = urlParams.get("id");
getUser(userID);

let editButton = document.querySelector("#editButton");
editButton.setAttribute("href", "/edit-user?id=" + userID);

let revealButton = document.querySelector("#revealButton");
revealButton.addEventListener("click", revealPaymentInputs);

let deactivateButton = document.querySelector("#deactivateButton");
deactivateButton.addEventListener("click", deactivateUser);

let activateButton = document.querySelector("#activateButton");
activateButton.addEventListener("click", activateUser);

let archiveButton = document.querySelector("#archiveButton");
archiveButton.addEventListener("click", archiveUser);

let restoreButton = document.querySelector("#restoreButton");
restoreButton.addEventListener("click", restoreUser);

let closePaymentModal = document.querySelector("#closePaymentModal");
closePaymentModal.addEventListener("click", () => {
  let paymentModal = document.querySelector("#paymentModal");
  paymentModal.classList.remove("is-active");
});

let body = document.querySelector("body");
body.addEventListener("keyup", event => {
  if (event.keyCode === 27) {
    let paymentModal = document.querySelector("#paymentModal");
    paymentModal.classList.remove("is-active");
  }
});

body.addEventListener("keyup", event => {
  if (event.keyCode === 13) {
    revealPaymentInputs();
  }
});

function deactivateUser() {
  fetch("users/" + userID + "/deactivate", {
    method: "POST",
  }).then(() => {
    window.location.replace("/user?id=" + userID);
  });
}

function activateUser() {
  fetch("users/" + userID + "/activate", {
    method: "POST",
  }).then(() => {
    window.location.replace("/user?id=" + userID);
  });
}

function replaceArchiveButtonToRestoreButton() {
  let archiveButtonGrandParent = archiveButton.parentElement.parentElement;
  archiveButtonGrandParent.setAttribute("hidden", "");

  let restoreButtonGrandParent = restoreButton.parentElement.parentElement;
  restoreButtonGrandParent.removeAttribute("hidden");
}

function replaceDeactivateButtonToActivateButton() {
  let deactivateButtonGrandParent = deactivateButton.parentElement.parentElement;
  deactivateButtonGrandParent.setAttribute("hidden", "");

  let activateButtonGrandParent = activateButton.parentElement.parentElement;
  activateButtonGrandParent.removeAttribute("hidden");
}

function hideRevealButton() {
  let revealButtonParent = revealButton.parentElement;
  revealButtonParent.setAttribute("hidden", "");
}

function hideDeactivateButton() {
  let deactivateButtonGrandParent = deactivateButton.parentElement.parentElement;
  deactivateButtonGrandParent.setAttribute("hidden", "");
}

function showPayments(payments) {
  for (let payment of payments) {
    let tr = document.createElement("tr");
    let adminTD = document.createElement("td");
    adminTD.append(payment.admin);
    let receiptTD = document.createElement("td");
    receiptTD.append(payment.receipt);
    let sumTD = document.createElement("td");
    sumTD.append(payment.sum);
    let dateTD = document.createElement("td");
    const d = new Date(payment.date);
    const date = d.getDate() + "." + (d.getMonth() + 1) + "." + d.getFullYear();
    dateTD.append(date);
    let tds = [];
    tds.push(adminTD, receiptTD, dateTD, sumTD);
    tr.append(...tds);
    document.querySelector("#tbody").append(tr);
  }
}

function showOperations(operations) {
  for (let operation of operations) {
    let tr = document.createElement("tr");
    let adminTD = document.createElement("td");
    adminTD.append(operation.admin);
    let actionTD = document.createElement("td");
    if (operation.type === "deactivate") {
      actionTD.append("Выключил");
    } else if (operation.type === "activate") {
      actionTD.append("Включил");
    }
    if (operation.admin === "ssn") {
      actionTD.append("а");
    }
    let dateTD = document.createElement("td");
    const d = new Date(operation.date);
    const date = d.getDate() + "." + (d.getMonth() + 1) + "." + d.getFullYear();
    dateTD.append(date);
    let tds = [];
    tds.push(adminTD, actionTD, dateTD);
    tr.append(...tds);
    document.querySelector("#operations").append(tr);
  }
}

function getUser(userID) {
  fetch("/users/" + userID)
    .then(res => {
      return res.json();
    })
    .then(user => {
      document.querySelector("#name").append(user.name);
      document.querySelector("#agreement").append(user.agreement);
      document.querySelector("#mac").append(user.mac);
      document.querySelector("#login").append(user.login);
      document.querySelector("#tariff").append(user.tariff.name);
      document.querySelector("#innerIP").append(user.inner_ip);
      document.querySelector("#extIP").append(user.ext_ip);
      document.querySelector("#phone").append(user.phone);
      document.querySelector("#room").append(user.room);
      document.querySelector("#comment").append(user.comment);
      document.querySelector("#connectionPlace").append(user.connection_place);

      if (user.is_employee === true) {
        document.querySelector("#isEmployee").append("Да");
      } else {
        document.querySelector("#isEmployee").append("Нет");
      }

      if (user.activity === true) {
        const d = new Date(user.expired_date);
        const expiredDate = d.getDate() + "." + (d.getMonth() + 1) + "." + d.getFullYear();
        document.querySelector("#expiredDate").append(expiredDate);
      } else {
        document.querySelector("#expiredDate").parentElement.remove();
        hideDeactivateButton();
      }
      document.querySelector("#balance").append(user.balance);

      if (user.is_archived) {
        replaceArchiveButtonToRestoreButton();
        hideRevealButton();
        hideDeactivateButton();
      }

      if (user.is_deactivated) {
        replaceDeactivateButtonToActivateButton();
      }

      if (user.payments !== undefined) {
        showPayments(user.payments);
      }

      if (user.operations !== undefined) {
        showOperations(user.operations);
      }
    });
}

function deposit() {
  fetch("payment", {
    method: "POST",
    headers: {
      "Content-Type": "application/json; charset=utf-8",
    },
    body: JSON.stringify({
      id: parseInt(userID),
      receipt: "№" + document.querySelector("#receipt").value + " от " + document.querySelector("#receiptDate").value,
      sum: parseInt(document.querySelector("#paymentSum").value),
    }),
  }).then(() => {
    location.replace("/user?id=" + userID);
  });
}

function revealPaymentInputs() {
  body.removeEventListener("keyup", event => {
    if (event.keyCode === 13) {
      revealPaymentInputs();
    }
  });

  let paymentModal = document.querySelector("#paymentModal");
  paymentModal.classList.add("is-active");

  paymentButton.addEventListener("click", deposit);
  let paymentSum = document.querySelector("#paymentSum");
  paymentSum.addEventListener("keyup", event => {
    if (event.keyCode === 13) {
      deposit();
    }
  });

  let receipt = document.querySelector("#receipt");
  receipt.focus();
}

function archiveUser() {
  let answer = confirm("Вы действительно хотите поместить этого пользователя в архив?");
  if (answer === true) {
    fetch("users/" + userID, {
      method: "DELETE",
    }).then(() => {
      window.location.replace("/");
    });
  }
}

function restoreUser() {
  let answer = confirm("Вы действительно хотите восстановить этого пользователя из архива?");
  if (answer === true) {
    fetch("users/" + userID, {
      method: "PUT",
    }).then(() => {
      window.location.replace("/");
    });
  }
}
