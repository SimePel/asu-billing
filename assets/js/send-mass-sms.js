function getUsers() {
    function createTD(tableTag, value) {
        let td = document.createElement("td");
        td.setAttribute("data-table-tag", tableTag);
        td.append(value);
        return td;
    }

    function createStatusTD(paid, activity) {
        let td = document.createElement("td");

        let paidText = "Не оплачено. ";
        if (paid) {
            paidText = "Оплачено. ";
        }

        let span = document.createElement("span");
        span.classList.add("icon", "has-text-danger");
        span.setAttribute("title", "Без доступа в интернет");
        let i = document.createElement("i");
        i.classList.add("fas", "fa-ban");
        if (activity) {
            i.classList.replace("fa-ban", "fa-check");
            span.classList.replace("has-text-danger", "has-text-success");
            span.setAttribute("title", "Подключен к интернету");
        }
        span.append(i);

        td.append(paidText);
        td.append(span);
        return td;
    }

    fetch("users")
        .then(res => {
            return res.json();
        })
        .then(users => {
            users.forEach(user => {
                if (user.phone === "") {
                    return;
                }
                let tds = [];

                tds.push(createTD("name", user.name));
                tds.push(createTD("phone", user.phone));
                tds.push(createStatusTD(user.paid, user.activity));

                let tr = document.createElement("tr");
                tr.setAttribute("data-is-sending", "false");
                if (user.activity) {
                    tr.setAttribute("data-is-sending", "true");
                }

                tr.setAttribute("data-is-active", "false");
                if (user.activity) {
                    tr.setAttribute("data-is-active", "true");
                }

                tr.classList.add("clickable");
                tr.addEventListener("click", () => {
                    if (tr.getAttribute("data-is-sending") === "true") {
                        tr.setAttribute("data-is-sending", "false");
                        return;
                    }
                    tr.setAttribute("data-is-sending", "true");
                });
                tr.append(...tds);

                document.getElementById("tbody").append(tr);
            });
        });
}

function setTrueSendingStatusForAllUsers() {
    document.querySelectorAll("tr.clickable").forEach((tr) => {
        tr.setAttribute("data-is-sending", "true");
    });
}

function setTrueSendingStatusForActiveUsers() {
    document.querySelectorAll("tr[data-is-active=\"true\"]").forEach((tr) => {
        tr.setAttribute("data-is-sending", "true");
    });

    document.querySelectorAll("tr[data-is-active=\"false\"]").forEach((tr) => {
        tr.setAttribute("data-is-sending", "false");
    });
}

function setTrueSendingStatusForInactiveUsers() {
    document.querySelectorAll("tr[data-is-active=\"false\"]").forEach((tr) => {
        tr.setAttribute("data-is-sending", "true");
    });

    document.querySelectorAll("tr[data-is-active=\"true\"]").forEach((tr) => {
        tr.setAttribute("data-is-sending", "false");
    });
}

getUsers();

document.querySelector("#all").addEventListener("click", setTrueSendingStatusForAllUsers);
document.querySelector("#active").addEventListener("click", setTrueSendingStatusForActiveUsers);
document.querySelector("#inactive").addEventListener("click", setTrueSendingStatusForInactiveUsers);