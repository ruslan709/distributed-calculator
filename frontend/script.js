document.addEventListener('DOMContentLoaded', function() {
    checkAuthentication();
});

function checkAuthentication() {
    const token = localStorage.getItem('jwt');
    if (token) {
        document.getElementById('auth-section').style.display = 'none';
        document.getElementById('app-section').style.display = 'block';
        // Initialize app functions
        // initializeApp();
    } else {
        document.getElementById('app-section').style.display = 'none';
        document.getElementById('auth-section').style.display = 'block';
    }
}

function login() {
    const username = document.getElementById('login-username').value;
    const password = document.getElementById('login-password').value;

    fetch('http://localhost:8080/api/v1/login', {
        method: 'POST',
        headers: {
            'Content-Type': 'application/json'
        },
        body: JSON.stringify({ Login: username, Password: password })
    })
    .then(response => {
        if (!response.ok) {
            throw new Error('Login failed');
        }
        return response.json();
    })
    .then(data => {
        localStorage.setItem('jwt', data.jwt);
        const id = getUserIDFromJWT();
        // console.log('id', id);
        checkAuthentication();
        fetchServerStatuses();
        loadAllCalculations(); // Загрузка всех вычислений при загрузке страницы
    })
    .catch(error => {
        console.error('Error logging in:', error);
        alert('Error logging in: ' + error.message);
    });
}


function switchToRegister() {
    document.getElementById('login-form').style.display = 'none';
    document.getElementById('register-form').style.display = 'block';
}

function switchToLogin() {
    document.getElementById('register-form').style.display = 'none';
    document.getElementById('login-form').style.display = 'block';
}



// Функция выхода
function logout() {
    localStorage.removeItem('jwt'); 
    document.getElementById('app-section').style.display = 'none';
    document.getElementById('auth-section').style.display = 'block';
    alert('You have been logged out.');
}

// Функция очистки результатов калькуляции
function clearOperations() {
    const calculationResultsSection = document.getElementById('calculation-results');
    if (calculationResultsSection) {
        calculationResultsSection.innerHTML = '';
    }
}

function getUserIDFromJWT() {
    const token = localStorage.getItem('jwt');
    if (!token) return null;

    const base64Url = token.split('.')[1]; // Access the payload part of the token
    const base64 = base64Url.replace(/-/g, '+').replace(/_/g, '/'); // Normalize base64 string
    const jsonPayload = decodeURIComponent(atob(base64).split('').map(function(c) {
        return '%' + ('00' + c.charCodeAt(0).toString(16)).slice(-2);
    }).join(''));

    const payload = JSON.parse(jsonPayload);
    return payload.userID; // Make sure the key matches the payload's key
}

// Функция очистки формы регистрации
function clearRegistrationForm() {
    document.getElementById('register-username').value = ''; // Clears the username input
    document.getElementById('register-password').value = ''; // Clears the password input
}

// Функция регистрации
function register() {
    const username = document.getElementById('register-username').value;
    console.log('username', username)
    const password = document.getElementById('register-password').value;
    console.log('password', password)

    fetch('http://localhost:8080/api/v1/register', {
        method: 'POST',
        headers: {
            'Content-Type': 'application/json'
        },
        body: JSON.stringify({ Login: username, Password: password })  // Make sure keys match server expectations
    })
    .then(response => response.json())  // Always expect JSON response
    .then(data => {
        if (data.success) {
            alert('Registration successful, please log in.');
            clearRegistrationForm();
        } else {
            throw new Error(data.error || 'Registration failed.');
        }
    })
    .catch(error => {
        console.error('Error registering:', error);
        alert(error.message);
    });
}

// function initializeApp() {
//     // Load calculations, statuses, etc.
//     fetchServerStatuses();
//     loadAllCalculations();
//     document.getElementById('reload-server-status').addEventListener('click', fetchServerStatuses);
//     document.getElementById('reload-operations-status').addEventListener('click', updateResults);
//     setInterval(updateResults, 60000);
//     applySettings();
//     document.querySelector('button[onclick="saveSettings()"]').addEventListener('click', applySettings);
// }

// Отправить вычисление на сервер
function submitCalculation() {
    const expression = document.getElementById('expression').value; // Получаем выражение от пользователя
    const calculationResultsSection = document.getElementById('calculation-results'); // Получаем секцию для вывода результатов

    // Проверяем выражение на валидность и на деление на ноль
    if (/\/0/.test(expression)) {
        appendCalculationResult(calculationResultsSection, null, `[${expression}] - Division by zero is not allowed.`, 'error');
        return;
    }
    // Улучшенное регулярное выражение для проверки валидности арифметического выражения
    if (!/^\d+([\+\-\*\/]\(?\-?\d+\)?)+$/.test(expression)) {
        appendCalculationResult(calculationResultsSection, null, `[${expression}] - Invalid expression format.`, 'error');
        return;
    }

    // Отправляем запрос на сервер
    fetch('http://localhost:8080/submit-calculation', {
        method: 'POST',
        headers: {
            'Content-Type': 'application/json'
        },
        body: JSON.stringify({
            UserId: getUserIDFromJWT(),
            operation: expression,
            add_duration: parseInt(document.getElementById('plus-time').value),
            subtract_duration: parseInt(document.getElementById('minus-time').value),
            multiply_duration: parseInt(document.getElementById('multiply-time').value),
            divide_duration: parseInt(document.getElementById('divide-time').value),
            inactive_server_time: parseInt(document.getElementById('inactive-server-time').value),
        })
    })
    .then(response => response.json())
    .then(data => {
        if (data.status === 'created') {
             // Сохраняем ID в localStorage
            let successfulIDs = JSON.parse(localStorage.getItem('successfulIDs')) || [];
            successfulIDs.push(data.id);
            localStorage.setItem('successfulIDs', JSON.stringify(successfulIDs));

            // Добавляем новый элемент для отображения операции и статуса
            appendCalculationResult(calculationResultsSection, data.id, data.operation, 'pending');
        } else if (data.status === 'error') {
            // Добавляем сообщение об ошибке и операцию
            appendCalculationResult(calculationResultsSection, `${expression} - ${data.error}`, 'error');
        }
    })
    .catch((error) => {
        console.error('Error:', error);
        appendCalculationResult(calculationResultsSection, `${expression} - Error: ${error}`, 'error');
    });
}

// // Функция для загрузки всех вычислений и предотвращения дубликатов
// function loadAllCalculations() {
//     // Очищаем существующие вычисления, чтобы избежать дублирования
//     const calculationResultsSection = document.getElementById('calculation-results');
//     calculationResultsSection.innerHTML = '';

//     fetch('http://localhost:8080/get-all-calculations')
//         .then(response => response.json())
//         .then(data => {
//             data.forEach(calculation => {
//                 const status = calculation.status === 'completed' ? 'success' : 'pending';
//                 const resultText = calculation.status === 'completed' ? calculation.result : '?';
//                 appendCalculationResult(calculationResultsSection, calculation.id, `${calculation.operation} Result = ${resultText}`, status);
//             });
//         })
//         .catch(error => console.error('Error loading calculations:', error));
// }

// Функция для загрузки всех вычислений конкретного пользователя по его userId
function loadAllCalculations() {
    const calculationResultsSection = document.getElementById('calculation-results');
    calculationResultsSection.innerHTML = '';  // Очищаем существующие вычисления, чтобы избежать дублирования

    const userId = getUserIDFromJWT(); // Получаем userId из JWT, хранящегося в localStorage
    if (!userId) {
        console.error('User ID is missing or invalid');
        return; // Возвращаем ошибку или прекращаем выполнение, если ID пользователя не найден
    }

    fetch(`http://localhost:8080/get-calculations-by-user?userId=${userId}`)
        .then(response => response.json())
        .then(data => {
            data.forEach(calculation => {
                const status = calculation.status === 'completed' ? 'success' : 'pending';
                const resultText = calculation.status === 'completed' ? calculation.result : '?';
                appendCalculationResult(calculationResultsSection, calculation.id, `${calculation.operation} Result = ${resultText}`, status);
            });
        })
        .catch(error => console.error('Error loading calculations:', error));
}

// Функция appendCalculationResult для динамического контента в зависимости от статуса
function appendCalculationResult(parentElement, id, message, status) {
    const resultElement = document.createElement('div');
    resultElement.className = `calculation-result ${status}`;
    resultElement.id = `result-${id}`;

    const idLine = document.createElement('div');
    idLine.textContent = `ID: ${id}`;
    resultElement.appendChild(idLine);

    const operationLine = document.createElement('div');
    operationLine.textContent = message;
    resultElement.appendChild(operationLine);

    if (status === 'pending') {
        const pendingLine = document.createElement('div');
        pendingLine.textContent = 'Expression will be calculated soon.';
        resultElement.appendChild(pendingLine);
    }

    parentElement.appendChild(resultElement);
}

// Сохранение настроек (в этом примере используется локальное хранилище)
function saveSettings() {
    localStorage.setItem('plus-time', document.getElementById('plus-time').value);
    localStorage.setItem('minus-time', document.getElementById('minus-time').value);
    localStorage.setItem('multiply-time', document.getElementById('multiply-time').value);
    localStorage.setItem('divide-time', document.getElementById('divide-time').value);
    localStorage.setItem('inactive-server-time', document.getElementById('inactive-server-time').value);

    alert('Settings saved successfully.');
}

// Функция для получения статуса серверов
function fetchServerStatuses() {
    const serverStatusesDiv = document.getElementById('server-statuses');
    // Очистка существующих статусов для исключения дубликатов
    serverStatusesDiv.innerHTML = '';

    // Получение статуса сервера orchestrator
    fetch('http://localhost:8080/orchestrator-status')
    .then(response => response.json())
    .then(orchestratorStatus => {
        // Отображение статус сервера orchestrator
        const orchestratorDiv = document.createElement('div');
        orchestratorDiv.className = orchestratorStatus.running ? 'server-status running' : 'server-status error';
        orchestratorDiv.innerHTML = `
            <p><strong>Type:</strong> Orchestrator</p>
            <p><strong>URL:</strong> http://localhost:8080/</p>
            <p><strong>Status:</strong> ${orchestratorStatus.running ? 'Running' : 'Not Running'} - ${orchestratorStatus.message}</p>
        `;
        serverStatusesDiv.appendChild(orchestratorDiv);
    })
    .catch(error => {
        console.error('Error fetching orchestrator status:', error);
        appendErrorServerDiv(serverStatusesDiv, 'Orchestrator', 'http://localhost:8080/', 'Unavailable - Could not connect');
    });

    // Получение статусов серверов calculator
    fetch('http://localhost:8080/ping-servers')
    .then(response => response.json())
    .then(calculatorStatuses => {
        calculatorStatuses.forEach(server => {
            appendServerStatusDiv(serverStatusesDiv, server);
        });
    })
    .catch(error => {
        console.error('Error fetching calculator servers statuses:', error);
        appendErrorServerDiv(serverStatusesDiv, 'Calculator', 'Unavailable URL', 'Unavailable - Could not connect');
    });
}

// Функция для добавления div элементов статусов серверов
function appendServerStatusDiv(parentElement, server) {
    const serverDiv = document.createElement('div');
    serverDiv.className = server.running ? 'server-status running' : 'server-status error';
    serverDiv.innerHTML = `
        <p><strong>Type:</strong> Calculator</p>
        <p><strong>URL:</strong> ${server.url}</p>
        <p><strong>Status:</strong> ${server.running ? 'Running' : 'Not Running'}${server.error ? ` (Error: ${server.error})` : ''}</p>
        ${server.running ? `<p><strong>Max Goroutines:</strong> ${server.maxGoroutines}</p>
        <p><strong>Current Goroutines:</strong> ${server.currentGoroutines}</p>` : ''}
    `;
    parentElement.appendChild(serverDiv);
}

// Функция для добавления div элементов ошибочных статусов серверов
function appendErrorServerDiv(parentElement, type, url, status) {
    const errorDiv = document.createElement('div');
    errorDiv.className = 'server-status error';
    errorDiv.innerHTML = `
        <p><strong>Type:</strong> ${type}</p>
        <p><strong>URL:</strong> ${url}</p>
        <p><strong>Status:</strong> ${status}</p>
    `;
    parentElement.appendChild(errorDiv);
}

// Функция для обновления результатов операций
function updateResults() {
    // Получение всех 'pending' операций на странице
    const pendingResults = document.querySelectorAll('.calculation-result.pending');

    pendingResults.forEach(resultElement => {
        const id = resultElement.id.split('-')[1]; // Предполагается, что формат ID - "result-{id}"

        // Запрашиваем результат операции по ID
        fetch(`http://localhost:8080/get-calculation-result?id=${id}`)
            .then(response => response.json())
            .then(data => {
                if (data.status === 'completed' && data.result !== undefined) {
                    // Обновляем текст результата и класс элемента
                    const operationLine = resultElement.querySelector('div:last-child');
                    operationLine.textContent = `[${data.operation}] Result = ${data.result}`;
                    resultElement.classList.remove('pending');
                    resultElement.classList.add('success');
                    resultElement.style.backgroundColor = "#4CAF50"; // Зеленый фон для завершенных операций
                } else {
                    // Если статус не завершен или результат отсутствует, оставляем как есть
                    console.log(`Calculation ID ${id} is still pending.`);
                }
            })
            .catch(error => console.error('Error updating result:', error));
    });
}

// Функция для очистки и обновления результатов операций
function clearAllCalculationsAndUpdate() {
    fetch('http://localhost:8080/clear-all-calculations', {
        method: 'POST',
        headers: {
            'Content-Type': 'application/json'
        }
    })
    .then(response => {
        if (response.ok) {
            console.log('All calculations cleared successfully.');
            // Очищаем отображаемые результаты
            document.getElementById('calculation-results').innerHTML = '';
            // По желанию, повторно загружаем и отображаем все вычисления
            updateResults();
        } else {
            console.error('Failed to clear calculations.');
        }
    })
    .catch(error => console.error('Error:', error));
}

document.addEventListener('DOMContentLoaded', function() {
    // Начальная настройка и обработчики событий
    fetchServerStatuses();
    document.getElementById('reload-server-status').addEventListener('click', fetchServerStatuses);
    document.getElementById('reload-operations-status').addEventListener('click', updateResults);
    loadAllCalculations(); // Загрузка всех вычислений при загрузке страницы

    // Автоматическое обновление всех вычислений каждую минуту
    setInterval(updateResults, 60000);

    // Применение настроек изначально и каждый раз при их сохранении
    applySettings();
    document.querySelector('button[onclick="saveSettings()"]').addEventListener('click', applySettings);
});

// Адаптация функции применения настроек и обработки блоков с ошибками сервера
function applySettings() {
    // Получение и применение настройки времени неактивности сервера
    const inactiveServerTime = parseInt(localStorage.getItem('inactive-server-time'), 10) || 60; // Значение по умолчанию 60 секунд
    localStorage.setItem('inactive-server-time', inactiveServerTime); // Обновление, если использовалось значение по умолчанию
    console.log(`Settings applied. Inactive server time: ${inactiveServerTime} seconds.`);

    // Настройка таймаута для очистки блоков с ошибками на основе времени неактивности сервера
    setTimeout(() => {
        const errorDivs = document.querySelectorAll('.server-status.error');
        errorDivs.forEach(div => div.remove());
        console.log('Old error divs removed based on inactive server time setting.');
    }, inactiveServerTime * 1000); // Перевод секунд в миллисекунды
}