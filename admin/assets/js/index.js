function getUsers() {
    function createTD(value) {
        let td = document.createElement("td");
        td.append(value);
        return td;
    }

    function createStatusIcon(activity) {
        let td = document.createElement("td");
        let span = document.createElement("span");
        let i = document.createElement("i");
        if (activity === true) {
            span.classList.add("icon", "has-text-success");
            i.classList.add("fas", "fa-user-check");
        } else {
            span.classList.add("icon", "has-text-danger");
            i.classList.add("fas", "fa-user-times");
        }
        span.append(i);
        td.append(span);

        return td;
    }

    fetch("users").then((res) => { return res.json() }).then((users) => {
        users.forEach(user => {
            let tds = [];

            tds.push(createTD(user.name));
            tds.push(createTD(user.agreement));
            tds.push(createTD(user.login));
            let expiredDate = "Не подключен";
            if (user.activity === true) {
                const d = new Date(user.expired_date);
                expiredDate = d.getDay() + "." + d.getMonth() + "." + d.getFullYear();
            }
            tds.push(createTD(expiredDate));
            tds.push(createTD(user.inner_ip));
            tds.push(createTD(user.phone));
            tds.push(createTD(user.room));
            tds.push(createTD(user.tariff.name));
            tds.push(createTD(user.connection_place));
            tds.push(createTD(user.balance));
            tds.push(createStatusIcon(user.activity));

            let tr = document.createElement("tr");
            tr.append(...tds);
            tr.classList.add("clickable");
            tr.addEventListener("click", (e) => {
                window.location.href = "/user?id=" + user.id;
            })
            document.getElementById("tbody").append(tr);
        });
    })
}

window.onload = () => {
    getUsers();

    /*Надеюсь верно разместил код, потому что  разбираться во всех js файлах жутко долго...
    
    Поскольку jQuery я не обнаружил,решил цепануть jQuery, он значительно упрощает такие вещи, думаю и в дальнейшем пригодится, но можно переписать и на чистый js

    Код при клике на элемент меню (li) меняем ему класс active, соответственно подцепляя необходимые стили, при повторном клике убираем этот класс...

    Вообще реализация этой штуки лучше, когда знаешь что происходит под капотом.. Там видимо выгружаются данные user'а и отображаются активные пункты меню, соответственно нужно будет накинуть класс active на активные пункты меню при выгрузке с БД или откуда-либо, чтобы активные отображались сразу, при загрузке страницы пользователем.. Думаю это ты реализуешь, если понял о чем я говорю..
    */
    $('.menu-item').on('click', function(){
        $(this).toggleClass('active');
    });

    /* Кнопка СМС оповещений будет отображать ВКЛ и зеленый BG Color если включены, и ВЫКЛ и красный, если выключено */
    
    $('.toggle').on('click', function(){
        $(this).toggleClass('active').find('a').text($(this).text() == 'Вкл' ? 'Выкл' : 'Вкл');
    });

    
}