async function fillAgreementField() {
  let response = await fetch('next-agreement');
  if (!response.ok) {
    alert('Ошибка HTTP: ' + response.status);
    return;
  }

  let json = await response.json();
  document.getElementsByName('agreement')[0].value = json.agreement;
}

window.addEventListener(
  'DOMContentLoaded',
  function () {
    document.getElementsByName('room')[0].addEventListener('input', () => {
      fetch('check-vacant-esockets', {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json; charset=utf-8',
        },
        body: JSON.stringify({
          room: document.getElementsByName('room')[0].value,
        }),
      })
        .then((res) => {
          return res.json();
        })
        .then((json) => {
          if (json.answer === true) {
            document.querySelector('#esocketsInfo').textContent = 'Есть свободная розетка';
          } else {
            document.querySelector('#esocketsInfo').textContent = 'Нет свободных розеток';
          }
        });
    });
  },
  false
);

fillAgreementField();
