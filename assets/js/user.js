const urlParams = new URLSearchParams(window.location.search);
const userID = urlParams.get("id");
getUser(userID);

let editButton = document.querySelector("#editButton");
editButton.setAttribute("href", "/edit-user?id=" + userID);

let paymentButton = document.querySelector("#paymentButton");
paymentButton.addEventListener("click", revealPaymentInputs);

let deactivateButton = document.querySelector("#deactivateButton");
deactivateButton.setAttribute("href", "/deactivate-user?id=" + userID);

let activateButton = document.querySelector("#activateButton");
activateButton.setAttribute("href", "/activate-user?id=" + userID);

let archiveButton = document.querySelector("#archiveButton");
archiveButton.addEventListener("click", archiveUser);

let restoreButton = document.querySelector("#restoreButton");
restoreButton.addEventListener("click", restoreUser);

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

function hidePaymentButton() {
  let paymentButtonGrandParent = paymentButton.parentElement.parentElement;
  paymentButtonGrandParent.setAttribute("hidden", "");
}

function hideDeactivateButton() {
  let deactivateButtonGrandParent = deactivateButton.parentElement.parentElement;
  deactivateButtonGrandParent.setAttribute("hidden", "");
}

function getUser(userID) {
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

  fetch("/users/" + userID)
    .then(res => {
      return res.json();
    })
    .then(user => {
      document.querySelector("#name").append(user.name);
      document.querySelector("#agreement").append(user.agreement);
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
      }
      document.querySelector("#balance").append(user.balance);

      if (user.is_archived) {
        replaceArchiveButtonToRestoreButton();
        hidePaymentButton();
        hideDeactivateButton();
      }

      if (user.is_stopped) {
        replaceDeactivateButtonToActivateButton();
      }

      if (user.payments !== undefined) {
        showPayments(user.payments);
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
      receipt: document.querySelector("#receiptInput").value,
      sum: parseInt(document.querySelector("#paymentInput").value),
    }),
  }).then(() => {
    location.replace("/user?id=" + userID);
  });
}

function revealPaymentInputs() {
  let receiptInput = document.querySelector("#receiptInput");
  let receiptInputParent = receiptInput.parentElement;

  receiptInputParent.removeAttribute("hidden");
  receiptInput.focus();

  let paymentInput = document.querySelector("#paymentInput");
  let paymentInputParent = paymentInput.parentElement;

  paymentInputParent.removeAttribute("hidden");
  paymentButton.removeEventListener("click", revealPaymentInputs);

  paymentButton.addEventListener("click", deposit);
  paymentInput.addEventListener("keyup", event => {
    if (event.keyCode === 13) {
      deposit();
    }
  });
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
