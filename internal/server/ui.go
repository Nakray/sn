package server
const indexHTML = `<!DOCTYPE html>
<html>
<head>
    <meta charset="utf-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
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
            <button class="tab active" onclick="showTab('tasks')">Monitoring Tasks</button>
            <button class="tab" onclick="showTab('accounts')">Accounts</button>
        </div>

        <!-- Tasks Tab -->
        <div id="tasks" class="tab-content active">
            <h2>Monitoring Tasks</h2>
            <form id="taskForm" onsubmit="createTask(event)">
                <div class="form-group">
                    <label>Owner Type:</label>
                    <select id="taskOwnerType" required>
                        <option value="user">User</option>
                        <option value="group">Group</option>
                    </select>
                </div>
                <div class="form-group">
                    <label>Owner ID:</label>
                    <input type="number" id="taskOwnerID" required>
                </div>
                <div class="form-group">
                    <label>Period (minutes):</label>
                    <input type="number" id="taskPeriod" value="60" required>
                </div>
                <div class="form-group">
                    <label>Account Group ID:</label>
                    <input type="number" id="taskAccountGroupID" value="0" required>
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
        <div id="accounts" class="tab-content">
            <h2>Accounts</h2>
            <form id="accountForm" onsubmit="createAccount(event)">
                <div class="form-group">
                    <label>Login:</label>
                    <input type="text" id="accountLogin" required>
                </div>
                <div class="form-group">
                    <label>Password:</label>
                    <input type="password" id="accountPassword" required>
                </div>
                <div class="form-group">
                    <label>Proxy (optional):</label>
                    <input type="text" id="accountProxy" placeholder="http://host:port">
                </div>
                <div class="form-group">
                    <label>Group ID:</label>
                    <input type="number" id="accountGroupID" value="0" required>
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
        function showTab(tabName) {
            const tabs = document.querySelectorAll('.tab-content');
            const buttons = document.querySelectorAll('.tab');
            tabs.forEach(tab => tab.classList.remove('active'));
            buttons.forEach(btn => btn.classList.remove('active'));
            document.getElementById(tabName).classList.add('active');
            event.target.classList.add('active');
        }
    </script>
</body>
</html>`
