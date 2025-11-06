package server

const indexHTML = `<!DOCTYPE html>
<html>
<head>
    <meta charset="utf-8" />
    <meta content="width=device-width, initial-scale=1.0" name="viewport" />
    <title>SN - VK Data Collector</title>
    <style>
        * { margin: 0; padding: 0; box-sizing: border-box; }
        body { font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', sans-serif; background: #f5f5f5; }
        .container { max-width: 1200px; margin: 0 auto; padding: 20px; }
        h1 { color: #333; margin-bottom: 30px; }
        h2 { color: #333; margin-bottom: 20px; }
        .tabs { display: flex; gap: 10px; margin-bottom: 20px; border-bottom: 2px solid #ddd; }
        .tab { padding: 10px 20px; cursor: pointer; background: none; border: none; font-size: 16px; color: #666; }
        .tab.active { color: #007bff; border-bottom: 2px solid #007bff; margin-bottom: -2px; }
        .tab-content { display: none; background: white; padding: 20px; border-radius: 8px; box-shadow: 0 2px 4px rgba(0,0,0,0.1); }
        .tab-content.active { display: block; }
        .form-group { margin-bottom: 15px; }
        label { display: block; margin-bottom: 5px; font-weight: 500; color: #333; }
        input, select { width: 100%; padding: 8px 12px; border: 1px solid #ddd; border-radius: 4px; font-size: 14px; }
        button { padding: 10px 20px; background: #007bff; color: white; border: none; border-radius: 4px; cursor: pointer; font-size: 14px; }
        button:hover { background: #0056b3; }
        button.danger { background: #dc3545; }
        button.danger:hover { background: #c82333; }
        table { width: 100%; border-collapse: collapse; margin-top: 20px; }
        th, td { padding: 12px; text-align: left; border-bottom: 1px solid #ddd; }
        th { background: #f8f9fa; font-weight: 600; color: #333; }
        .actions { display: flex; gap: 10px; }
        .status { padding: 4px 8px; border-radius: 4px; font-size: 12px; }
        .status.active { background: #d4edda; color: #155724; }
        .status.blocked { background: #f8d7da; color: #721c24; }
        .btn-small { padding: 4px 8px; font-size: 12px; }
    </style>
</head>
<body>
    <div class="container">
        <h1>SN - VK Data Collector</h1>
        <div class="tabs">
            <button class="tab active" onclick="showTab('tasks', this)">Monitoring Tasks</button>
            <button class="tab" onclick="showTab('accounts', this)">Accounts</button>
        </div>
        <!-- Tasks Tab -->
        <div class="tab-content active" id="tasks">
            <h2>Monitoring Tasks</h2>
            <form id="taskForm" onsubmit="createTask(event)">
                <div class="form-group">
                    <label>Owner Type:</label>
                    <select id="taskOwnerType" required="">
                        <option value="user">User</option>
                        <option value="group">Group</option>
                    </select>
                </div>
                <div class="form-group">
                    <label>Owner ID:</label>
                    <input id="taskOwnerID" required="" type="number"/>
                </div>
                <div class="form-group">
                    <label>Period (minutes):</label>
                    <input id="taskPeriod" required="" type="number" value="60"/>
                </div>
                <div class="form-group">
                    <label>Account Group ID:</label>
                    <input id="taskAccountGroupID" required="" type="number" value="0"/>
                </div>
                <button type="submit">Create Task</button>
            </form>
            <table id="tasksTable">
                <thead>
                    <tr>
                        <th>ID</th>
                        <th>Type</th>
                        <th>Owner ID</th>
                        <th>Period (min)</th>
                        <th>Last Run</th>
                        <th>Actions</th>
                    </tr>
                </thead>
                <tbody></tbody>
            </table>
        </div>
        <!-- Accounts Tab -->
        <div class="tab-content" id="accounts">
            <h2>Accounts</h2>
            <form id="accountForm" onsubmit="createAccount(event)">
                <div class="form-group">
                    <label>Login:</label>
                    <input id="accountLogin" required="" type="text"/>
                </div>
                <div class="form-group">
                    <label>Password:</label>
                    <input id="accountPassword" required="" type="password"/>
                </div>
                <div class="form-group">
                    <label>Proxy (optional):</label>
                    <input id="accountProxy" placeholder="http://host:port" type="text"/>
                </div>
                <div class="form-group">
                    <label>Group ID:</label>
                    <input id="accountGroupID" required="" type="number" value="0"/>
                </div>
                <button type="submit">Add Account</button>
            </form>
            <table id="accountsTable">
                <thead>
                    <tr>
                        <th>ID</th>
                        <th>Login</th>
                        <th>Proxy</th>
                        <th>Group ID</th>
                        <th>Status</th>
                        <th>Actions</th>
                    </tr>
                </thead>
                <tbody></tbody>
            </table>
        </div>
    </div>
    <script>
        function showTab(tabName, btn) {
            document.querySelectorAll('.tab').forEach(t => t.classList.remove('active'));
            document.querySelectorAll('.tab-content').forEach(c => c.classList.remove('active'));
            btn.classList.add('active');
            document.getElementById(tabName).classList.add('active');
        }

        async function loadTasks() {
            const res = await fetch('/api/tasks');
            const tasks = await res.json();
            const tbody = document.querySelector('#tasksTable tbody');
            tbody.innerHTML = tasks.map(function(t){
                return "<tr>" +
                    "<td>" + t.ID + "</td>" +
                    "<td>" + t.OwnerType + "</td>" +
                    "<td>" + t.OwnerID + "</td>" +
                    "<td>" + t.Period + "</td>" +
                    "<td>" + new Date(t.LastTimestamp).toLocaleString() + "</td>" +
                    "<td><button class=\"btn-small danger\" onclick=\"deleteTask(" + t.ID + ")\">Delete</button></td>" +
                "</tr>";
            }).join('');
        }

        async function createTask(e) {
            e.preventDefault();
            const task = {
                SocialNetworkType: 'vkontakte',
                OwnerType: document.getElementById('taskOwnerType').value,
                OwnerID: parseInt(document.getElementById('taskOwnerID').value),
                Period: parseInt(document.getElementById('taskPeriod').value),
                AccountGroupID: parseInt(document.getElementById('taskAccountGroupID').value),
                Filters: {},
                FilterLimits: {}
            };
            await fetch('/api/tasks', {
                method: 'POST',
                headers: {'Content-Type': 'application/json'},
                body: JSON.stringify(task)
            });
            document.getElementById('taskForm').reset();
            loadTasks();
        }

        async function deleteTask(id) {
            if (!confirm('Delete this task?')) return;
            await fetch('/api/tasks/' + id, {method: 'DELETE'});
            loadTasks();
        }

        async function loadAccounts() {
            const res = await fetch('/api/accounts');
            const accounts = await res.json();
            const tbody = document.querySelector('#accountsTable tbody');
            tbody.innerHTML = accounts.map(function(a){
                return "<tr>" +
                    "<td>" + a.ID + "</td>" +
                    "<td>" + a.Login + "</td>" +
                    "<td>" + (a.Proxy || '-') + "</td>" +
                    "<td>" + a.GroupID + "</td>" +
                    "<td><span class=\"status " + (a.IsBlocked ? "blocked" : "active") + "\">" + (a.IsBlocked ? "Blocked" : "Active") + "</span></td>" +
                    "<td><button class=\"btn-small danger\" onclick=\"deleteAccount(" + a.ID + ")\">Delete</button></td>" +
                "</tr>";
            }).join('');
        }

        async function createAccount(e) {
            e.preventDefault();
            const account = {
                SocialNetworkType: 'VKontakte',
                Login: document.getElementById('accountLogin').value,
                Password: document.getElementById('accountPassword').value,
                Proxy: document.getElementById('accountProxy').value,
                GroupID: parseInt(document.getElementById('accountGroupID').value),
                IsBlocked: false
            };
            await fetch('/api/accounts', {
                method: 'POST',
                headers: {'Content-Type': 'application/json'},
                body: JSON.stringify(account)
            });
            document.getElementById('accountForm').reset();
            loadAccounts();
        }

        async function deleteAccount(id) {
            if (!confirm('Delete this account?')) return;
            await fetch('/api/accounts/' + id, {method: 'DELETE'});
            loadAccounts();
        }

        // Load data on page load
        loadTasks();
        loadAccounts();
    </script>
</body>
</html>
`
