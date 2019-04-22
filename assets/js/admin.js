function displayTable() {
    let toDisplay = JSON.parse(localStorage.getItem('elems'));
    let defaultTable = ["name", "login", "tariff", "balance", "active"];

    if (toDisplay === null) {
        localStorage.setItem('elems', JSON.stringify(defaultTable));
        toDisplay = defaultTable;
    }

    for (let i in toDisplay) {
        document.querySelectorAll('.' + toDisplay[i]).forEach(function (td) {
            td.classList.remove("invisible");
        });
        document.getElementById(toDisplay[i] + 'Box').checked = true;
    }
}

window.onload = function () {
    displayTable();
    let url = new URL(window.location.href);
    let type = url.searchParams.get("type");
    let name = url.searchParams.get("name");
    let account = url.searchParams.get("account");

    if (type === "wired") {
        document.getElementById("wired").innerHTML = "<strong>Проводные</strong/";
    } else if (type === "wireless") {
        document.getElementById("wireless").innerHTML = "<strong>Беспроводные</strong/";
    } else if (type === "active") {
        document.getElementById("active").innerHTML = "<strong>Включенные</strong/";
    } else if (type === "inactive") {
        document.getElementById("inactive").innerHTML = "<strong>Отключенные</strong/";
    } else if ((name === null) && (account === null)) {
        document.getElementById("all").innerHTML = "<strong>Все</strong/";
    }

    const searchButton = document.getElementById("searchButton");
    const searchInput = document.getElementById("search");

    searchButton.addEventListener('click', search);

    searchInput.addEventListener('keyup', function (event) {
        if (event.keyCode === 13) {
            search();
        }
    });

    const dropdownTrigger = document.getElementById("dropdownTrigger");

    dropdownTrigger.addEventListener('click', function (event) {
        event.stopPropagation();
        document.getElementById("dropdown").classList.toggle('is-active');
    })

    const nameBox = document.getElementById("nameBox");
    const agreementBox = document.getElementById("agreementBox");
    const loginBox = document.getElementById("loginBox");
    const expiredDateBox = document.getElementById("expiredDateBox");
    const ipBox = document.getElementById("ipBox");
    const phoneBox = document.getElementById("phoneBox");
    const commentBox = document.getElementById("commentBox");
    const tariffBox = document.getElementById("tariffBox");
    const connectionPlaceBox = document.getElementById("connectionPlaceBox");
    const balanceBox = document.getElementById("balanceBox");
    const activeBox = document.getElementById("activeBox");

    nameBox.addEventListener('click', function (event) {
        let currentTable = JSON.parse(localStorage.getItem('elems'));
        if (nameBox.checked) {
            document.querySelectorAll('.name').forEach(function (td) {
                td.classList.remove("invisible");
            });
            currentTable.push('name');
        } else {
            document.querySelectorAll('.name').forEach(function (td) {
                td.classList.add("invisible");
            });
            currentTable = currentTable.filter(function (value, index, arr) {
                return value !== 'name';
            });
        }
        localStorage.setItem('elems', JSON.stringify(currentTable));
    })

    agreementBox.addEventListener('click', function (event) {
        let currentTable = JSON.parse(localStorage.getItem('elems'));
        if (agreementBox.checked) {
            document.querySelectorAll('.agreement').forEach(function (td) {
                td.classList.remove("invisible");
            });
            currentTable.push('agreement');
        } else {
            document.querySelectorAll('.agreement').forEach(function (td) {
                td.classList.add("invisible");
            });
            currentTable = currentTable.filter(function (value, index, arr) {
                return value !== 'agreement';
            });
        }
        localStorage.setItem('elems', JSON.stringify(currentTable));
    })

    loginBox.addEventListener('click', function (event) {
        let currentTable = JSON.parse(localStorage.getItem('elems'));
        if (loginBox.checked) {
            document.querySelectorAll('.login').forEach(function (td) {
                td.classList.remove("invisible");
            });
            currentTable.push('login');
        } else {
            document.querySelectorAll('.login').forEach(function (td) {
                td.classList.add("invisible");
            });
            currentTable = currentTable.filter(function (value, index, arr) {
                return value !== 'login';
            });
        }
        localStorage.setItem('elems', JSON.stringify(currentTable));
    })

    expiredDateBox.addEventListener('click', function (event) {
        let currentTable = JSON.parse(localStorage.getItem('elems'));
        if (expiredDateBox.checked) {
            document.querySelectorAll('.expiredDate').forEach(function (td) {
                td.classList.remove("invisible");
            });
            currentTable.push('expiredDate');
        } else {
            document.querySelectorAll('.expiredDate').forEach(function (td) {
                td.classList.add("invisible");
            });
            currentTable = currentTable.filter(function (value, index, arr) {
                return value !== 'expiredDate';
            });
        }
        localStorage.setItem('elems', JSON.stringify(currentTable));
    })

    ipBox.addEventListener('click', function (event) {
        let currentTable = JSON.parse(localStorage.getItem('elems'));
        if (ipBox.checked) {
            document.querySelectorAll('.ip').forEach(function (td) {
                td.classList.remove("invisible");
            });
            currentTable.push('ip');
        } else {
            document.querySelectorAll('.ip').forEach(function (td) {
                td.classList.add("invisible");
            });
            currentTable = currentTable.filter(function (value, index, arr) {
                return value !== 'ip';
            });
        }
        localStorage.setItem('elems', JSON.stringify(currentTable));
    })

    phoneBox.addEventListener('click', function (event) {
        let currentTable = JSON.parse(localStorage.getItem('elems'));
        if (phoneBox.checked) {
            document.querySelectorAll('.phone').forEach(function (td) {
                td.classList.remove("invisible");
            });
            currentTable.push('phone');
        } else {
            document.querySelectorAll('.phone').forEach(function (td) {
                td.classList.add("invisible");
            });
            currentTable = currentTable.filter(function (value, index, arr) {
                return value !== 'phone';
            });
        }
        localStorage.setItem('elems', JSON.stringify(currentTable));
    })

    commentBox.addEventListener('click', function (event) {
        let currentTable = JSON.parse(localStorage.getItem('elems'));
        if (commentBox.checked) {
            document.querySelectorAll('.comment').forEach(function (td) {
                td.classList.remove("invisible");
            });
            currentTable.push('comment');
        } else {
            document.querySelectorAll('.comment').forEach(function (td) {
                td.classList.add("invisible");
            });
            currentTable = currentTable.filter(function (value, index, arr) {
                return value !== 'comment';
            });
        }
        localStorage.setItem('elems', JSON.stringify(currentTable));
    })

    tariffBox.addEventListener('click', function (event) {
        let currentTable = JSON.parse(localStorage.getItem('elems'));
        if (tariffBox.checked) {
            document.querySelectorAll('.tariff').forEach(function (td) {
                td.classList.remove("invisible");
            });
            currentTable.push('tariff');
        } else {
            document.querySelectorAll('.tariff').forEach(function (td) {
                td.classList.add("invisible");
            });
            currentTable = currentTable.filter(function (value, index, arr) {
                return value !== 'tariff';
            });
        }
        localStorage.setItem('elems', JSON.stringify(currentTable));
    })

    connectionPlaceBox.addEventListener('click', function (event) {
        let currentTable = JSON.parse(localStorage.getItem('elems'));
        if (connectionPlaceBox.checked) {
            document.querySelectorAll('.connectionPlace').forEach(function (td) {
                td.classList.remove("invisible");
            });
            currentTable.push('connectionPlace');
        } else {
            document.querySelectorAll('.connectionPlace').forEach(function (td) {
                td.classList.add("invisible");
            });
            currentTable = currentTable.filter(function (value, index, arr) {
                return value !== 'connectionPlace';
            });
        }
        localStorage.setItem('elems', JSON.stringify(currentTable));
    })

    balanceBox.addEventListener('click', function (event) {
        let currentTable = JSON.parse(localStorage.getItem('elems'));
        if (balanceBox.checked) {
            document.querySelectorAll('.balance').forEach(function (td) {
                td.classList.remove("invisible");
            });
            currentTable.push('balance');
        } else {
            document.querySelectorAll('.balance').forEach(function (td) {
                td.classList.add("invisible");
            });
            currentTable = currentTable.filter(function (value, index, arr) {
                return value !== 'balance';
            });
        }
        localStorage.setItem('elems', JSON.stringify(currentTable));
    })

    activeBox.addEventListener('click', function (event) {
        let currentTable = JSON.parse(localStorage.getItem('elems'));
        if (activeBox.checked) {
            document.querySelectorAll('.active').forEach(function (td) {
                td.classList.remove("invisible");
            });
            currentTable.push('active');
        } else {
            document.querySelectorAll('.active').forEach(function (td) {
                td.classList.add("invisible");
            });
            currentTable = currentTable.filter(function (value, index, arr) {
                return value !== 'active';
            });
        }
        localStorage.setItem('elems', JSON.stringify(currentTable));
    })

    document.getElementById("activeUsers").innerText = document.getElementsByClassName("activeUsers").length;
    document.getElementById("inactiveUsers").innerText = document.getElementsByClassName("inactiveUsers").length;

};

function search() {
    var s = document.getElementById("select").value;
    if (s === "name") {
        window.location.replace("/adm?name=" + document.getElementById("search").value);
    } else if (s === "account") {
        window.location.replace("/adm?account=" + document.getElementById("search").value);
    } else {
        alert("Неопознанный тип поиска");
    }
}
