import {boardURL, onRouteChange, readRoute, setBoardRoute} from "/router.js";

const initialRoute = readRoute();
const userId = initialRoute.userId;
const selectedBoardStorageKey = `taskflow-selected-board-${userId}`;
const columnColors = ["#9b9dac", "#7160e8", "#e4a348", "#43ad83", "#de6f8c", "#4e9bc7"];
const priorityValues = {high: 10, medium: 50, low: 90};
const taskDragType = "application/x-taskflow-task";
const columnDragType = "application/x-taskflow-column";

const state = {
    boards: [],
    activeBoard: null,
    tasks: []
};

let activeTaskId = null;
let editingColumnId = null;
let confirmAction = null;
let toastTimer;

const board = document.querySelector("#kanban-board");
const boardList = document.querySelector("#board-list");
const searchInput = document.querySelector("#task-search");
const priorityFilter = document.querySelector("#priority-filter");
const drawerLayer = document.querySelector("#drawer-layer");
const drawerContent = document.querySelector("#drawer-content");
const drawerTaskId = document.querySelector("#drawer-task-id");
const taskDialog = document.querySelector("#task-dialog");
const taskForm = document.querySelector("#task-form");
const boardDialog = document.querySelector("#board-dialog");
const boardForm = document.querySelector("#board-form");
const columnDialog = document.querySelector("#column-dialog");
const columnForm = document.querySelector("#column-form");
const confirmDialog = document.querySelector("#confirm-dialog");
const confirmForm = document.querySelector("#confirm-form");
const toast = document.querySelector("#toast");

function escapeHTML(value = "") {
    return String(value)
        .replaceAll("&", "&amp;")
        .replaceAll("<", "&lt;")
        .replaceAll(">", "&gt;")
        .replaceAll('"', "&quot;")
        .replaceAll("'", "&#039;");
}

async function api(path, options = {}) {
    const headers = {...options.headers};
    if (options.body) headers["Content-Type"] = "application/json";
    const response = await fetch(`/api${path}`, {...options, headers});
    if (!response.ok) {
        const message = (await response.text()).trim();
        throw new Error(message || `Ошибка запроса: ${response.status}`);
    }
    if (response.status === 204) return null;
    const contentType = response.headers.get("content-type") ?? "";
    return contentType.includes("application/json") ? response.json() : null;
}

function columns() {
    return [...(state.activeBoard?.columns ?? [])].sort((left, right) => left.position - right.position || left.id - right.id);
}

function backlogColumn() {
    const availableColumns = columns().filter((column) => column.id !== state.activeBoard?.done_column_id);
    return availableColumns.find((column) => column.title.trim().toLocaleLowerCase("ru") === "backlog") ?? availableColumns[0];
}

function priorityKey(value) {
    if (value <= 33) return "high";
    if (value <= 66) return "medium";
    return "low";
}

function priorityData(value) {
    const priorities = {
        high: {label: "Высокий", color: "#e05268"},
        medium: {label: "Средний", color: "#f0a247"},
        low: {label: "Низкий", color: "#45a57f"}
    };
    return priorities[priorityKey(value)];
}

function dateInputValue(value) {
    if (!value) return "";
    const date = new Date(value);
    if (Number.isNaN(date.getTime())) return "";
    return date.toISOString().slice(0, 10);
}

function formatDate(value) {
    if (!value) return "Без даты";
    const date = new Date(value);
    if (Number.isNaN(date.getTime())) return "Без даты";
    return new Intl.DateTimeFormat("ru-RU", {day: "numeric", month: "short"}).format(date);
}

function durationMinutes(task) {
    return Math.round(Number(task.duration || 0) / 60_000_000_000);
}

function filteredTasks() {
    const text = searchInput.value.trim().toLocaleLowerCase("ru");
    const priority = priorityFilter.value;
    return state.tasks.filter((task) => {
        const matchesSearch = !text || `${task.title} ${task.description}`.toLocaleLowerCase("ru").includes(text);
        return matchesSearch && (priority === "all" || priorityKey(task.priority) === priority);
    });
}

async function loadBoards(preferredBoardId = null, replaceRoute = true) {
    state.boards = await api(`/boards?user_id=${userId}`);
    if (!state.boards.length) {
        state.activeBoard = null;
        state.tasks = [];
        setBoardRoute(userId, null, true);
        render();
        return;
    }

    const storedBoardId = Number(localStorage.getItem(selectedBoardStorageKey));
    const targetId = preferredBoardId || storedBoardId;
    state.activeBoard = state.boards.find((item) => item.id === targetId) ?? state.boards[0];
    localStorage.setItem(selectedBoardStorageKey, String(state.activeBoard.id));
    setBoardRoute(userId, state.activeBoard.id, replaceRoute);
    await loadTasks();
    render();
}

async function refreshActiveBoard() {
    if (!state.activeBoard) return;
    const refreshed = await api(`/boards/${state.activeBoard.id}`);
    state.activeBoard = refreshed;
    const index = state.boards.findIndex((item) => item.id === refreshed.id);
    if (index >= 0) state.boards[index] = refreshed;
    await loadTasks();
    render();
}

async function loadTasks() {
    if (!state.activeBoard) {
        state.tasks = [];
        return;
    }
    state.tasks = await api(`/tasks?user_id=${userId}&board_id=${state.activeBoard.id}`);
    state.tasks.sort((left, right) => left.position - right.position || left.id - right.id);
}

function render() {
    renderBoardList();
    renderHeader();
    renderBoard();
}

function renderBoardList() {
    boardList.innerHTML = state.boards.map((item, index) => `
        <a class="board-link ${item.id === state.activeBoard?.id ? "active" : ""}" href="${boardURL(userId, item.id)}" data-board-id="${item.id}">
            <span class="board-dot" style="background:${columnColors[index % columnColors.length]}"></span>
            <span>${escapeHTML(item.title)}</span>
        </a>
    `).join("");

    boardList.querySelectorAll("[data-board-id]").forEach((button) => {
        button.addEventListener("click", async (event) => {
            event.preventDefault();
            const next = state.boards.find((item) => item.id === Number(button.dataset.boardId));
            if (!next || next.id === state.activeBoard?.id) return;
            state.activeBoard = next;
            localStorage.setItem(selectedBoardStorageKey, String(next.id));
            setBoardRoute(userId, next.id);
            closeDrawer();
            await runAction(loadTasks, "Не удалось открыть доску");
            render();
        });
    });
}

function renderHeader() {
    const title = state.activeBoard?.title ?? "Создайте первую доску";
    const boardColumns = columns();
    document.querySelector("#board-eyebrow").textContent = state.activeBoard ? `Доска · пользователь ${userId}` : "Рабочее пространство";
    document.querySelector("#board-title").textContent = title;
    document.querySelector("#board-description").textContent = state.activeBoard
        ? `${boardColumns.length} колонок · ${state.tasks.length} задач`
        : "У одной учётной записи может быть несколько независимых досок.";
    document.querySelector("#profile-avatar").textContent = `U${userId}`;
    document.querySelector("#profile-name").textContent = `Пользователь ${userId}`;
    document.querySelector(".brand").href = boardURL(userId, state.activeBoard?.id);
    document.querySelector("#new-task-top").disabled = !state.activeBoard;
    document.querySelector("#new-column").disabled = !state.activeBoard;
}

function renderBoard() {
    if (!state.activeBoard) {
        board.innerHTML = `
            <div class="empty-board">
                <div class="empty-board-mark">T</div>
                <h2>Пока нет ни одной доски</h2>
                <p>Создайте доску — колонки Backlog, «В работе», «Проверка» и постоянная «Готово» появятся автоматически.</p>
                <button class="primary-button" type="button" data-create-first-board>Создать доску</button>
            </div>
        `;
        board.querySelector("[data-create-first-board]").addEventListener("click", createBoard);
        updateSummary();
        return;
    }

    const visibleTasks = filteredTasks();
    board.innerHTML = columns().map((column, index) => {
        const columnTasks = visibleTasks.filter((task) => task.column_id === column.id);
        const cards = columnTasks.length
            ? columnTasks.map(renderCard).join("")
            : '<div class="empty-column">Перетащите задачу сюда</div>';
        const systemClass = column.id === state.activeBoard.done_column_id ? " system-column" : "";
        return `
            <section class="kanban-column${systemClass}" data-column-id="${column.id}" style="--column-color:${columnColors[index % columnColors.length]}">
                <header class="column-header" draggable="true" data-column-drag="${column.id}" title="Перетащите, чтобы изменить позицию">
                    <span class="column-grip" aria-hidden="true">⋮⋮</span>
                    <span class="column-color"></span>
                    <h2>${escapeHTML(column.title)}</h2>
                    <span class="column-count">${columnTasks.length}</span>
                    <button class="column-menu" type="button" data-column-menu="${column.id}" aria-label="Меню колонки">•••</button>
                </header>
                <div class="column-tasks">${cards}</div>
            </section>
        `;
    }).join("");

    bindBoardEvents();
    updateSummary();
}

function renderCard(task) {
    const priority = priorityData(task.priority);
    const completed = task.is_done ? '<span class="task-label done">Готово</span>' : "";
    const minutes = durationMinutes(task);
    return `
        <article class="task-card" draggable="true" data-task-id="${task.id}" style="--priority-color:${priority.color}" tabindex="0">
            <div class="card-topline">
                <span class="task-code">TASK-${task.id}</span>
                <span class="priority-label" style="color:${priority.color}">${priority.label}</span>
            </div>
            <h3>${escapeHTML(task.title)}</h3>
            <p class="card-description">${escapeHTML(task.description || "Без описания")}</p>
            <div class="card-labels">${completed}</div>
            <div class="progress-row" aria-label="Прогресс ${task.progress}%">
                <div class="progress-track"><div class="progress-fill" style="width:${task.progress}%"></div></div>
                <span class="progress-value">${task.progress}%</span>
            </div>
            <footer class="card-footer">
                <span class="card-meta">◷ ${formatDate(task.start_time)}</span>
                ${minutes ? `<span class="card-meta">${minutes} мин</span>` : ""}
                <span class="assignee">U${task.user_id}</span>
            </footer>
        </article>
    `;
}

function bindBoardEvents() {
    document.querySelectorAll(".task-card").forEach((card) => {
        card.addEventListener("click", () => openDrawer(Number(card.dataset.taskId)));
        card.addEventListener("keydown", (event) => {
            if (event.key === "Enter" || event.key === " ") {
                event.preventDefault();
                openDrawer(Number(card.dataset.taskId));
            }
        });
        card.addEventListener("dragstart", (event) => {
            event.dataTransfer.effectAllowed = "move";
            event.dataTransfer.setData(taskDragType, card.dataset.taskId);
            requestAnimationFrame(() => card.classList.add("dragging"));
        });
        card.addEventListener("dragend", () => {
            card.classList.remove("dragging");
            clearColumnDragStyles();
        });
    });

    document.querySelectorAll("[data-column-drag]").forEach((header) => {
        header.addEventListener("dragstart", (event) => {
            if (event.target.closest("button")) {
                event.preventDefault();
                return;
            }
            event.dataTransfer.effectAllowed = "move";
            event.dataTransfer.setData(columnDragType, header.dataset.columnDrag);
            requestAnimationFrame(() => header.closest(".kanban-column").classList.add("column-dragging"));
        });
        header.addEventListener("dragend", clearColumnDragStyles);
    });

    document.querySelectorAll(".kanban-column").forEach((column) => {
        column.addEventListener("dragover", (event) => {
            const types = Array.from(event.dataTransfer.types);
            if (!types.includes(taskDragType) && !types.includes(columnDragType)) return;
            event.preventDefault();
            event.dataTransfer.dropEffect = "move";
            if (types.includes(columnDragType)) {
                const bounds = column.getBoundingClientRect();
                column.classList.add("column-drag-over");
                column.classList.toggle("drop-after", event.clientX > bounds.left + bounds.width / 2);
                return;
            }
            column.classList.add("task-drag-over");
        });
        column.addEventListener("dragleave", (event) => {
            if (!column.contains(event.relatedTarget)) {
                column.classList.remove("task-drag-over", "column-drag-over", "drop-after");
            }
        });
        column.addEventListener("drop", async (event) => {
            event.preventDefault();
            const columnId = Number(column.dataset.columnId);
            if (Array.from(event.dataTransfer.types).includes(columnDragType)) {
                const draggedColumnId = Number(event.dataTransfer.getData(columnDragType));
                const placeAfter = column.classList.contains("drop-after");
                clearColumnDragStyles();
                await reorderColumn(draggedColumnId, columnId, placeAfter);
                return;
            }
            column.classList.remove("task-drag-over");
            const task = state.tasks.find((item) => item.id === Number(event.dataTransfer.getData(taskDragType)));
            if (!task || task.column_id === columnId) return;
            await moveTask(task, columnId);
        });
    });

    document.querySelectorAll("[data-column-menu]").forEach((button) => {
        button.addEventListener("click", () => openColumnMenu(Number(button.dataset.columnMenu)));
    });
}

function clearColumnDragStyles() {
    document.querySelectorAll(".kanban-column").forEach((item) => {
        item.classList.remove("task-drag-over", "column-dragging", "column-drag-over", "drop-after");
    });
}

function updateSummary() {
    const progress = state.tasks.length
        ? Math.round(state.tasks.reduce((sum, task) => sum + task.progress, 0) / state.tasks.length)
        : 0;
    document.querySelector("#summary-progress").textContent = `${progress}%`;
    document.querySelector("#summary-ring").style.setProperty("--progress", `${progress * 3.6}deg`);
}

function openDrawer(taskId) {
    const task = state.tasks.find((item) => item.id === taskId);
    if (!task) return;
    activeTaskId = taskId;
    drawerTaskId.textContent = `TASK-${task.id}`;
    drawerContent.innerHTML = renderDrawer(task);
    bindDrawerEvents(task);
    drawerLayer.classList.add("open");
    drawerLayer.setAttribute("aria-hidden", "false");
}

function closeDrawer() {
    drawerLayer.classList.remove("open");
    drawerLayer.setAttribute("aria-hidden", "true");
    activeTaskId = null;
}

function renderDrawer(task) {
    const selectedPriority = priorityKey(task.priority);
    return `
        <input class="drawer-title-input" id="drawer-title" value="${escapeHTML(task.title)}" maxlength="100" aria-label="Название задачи">
        <textarea class="drawer-description" id="drawer-description" aria-label="Описание задачи">${escapeHTML(task.description)}</textarea>
        <label class="completion-field">
            <input id="drawer-done" type="checkbox" ${task.is_done ? "checked" : ""}>
            <span><strong>Задача выполнена</strong><small>Переместить в системную колонку «Готово»</small></span>
        </label>
        <div class="detail-grid">
            <label class="detail-field">
                <span>Колонка</span>
                <select id="drawer-column">
                    ${columns().map((column) => `<option value="${column.id}" ${column.id === task.column_id ? "selected" : ""}>${escapeHTML(column.title)}</option>`).join("")}
                </select>
            </label>
            <label class="detail-field">
                <span>Приоритет</span>
                <select id="drawer-priority">
                    <option value="high" ${selectedPriority === "high" ? "selected" : ""}>Высокий</option>
                    <option value="medium" ${selectedPriority === "medium" ? "selected" : ""}>Средний</option>
                    <option value="low" ${selectedPriority === "low" ? "selected" : ""}>Низкий</option>
                </select>
            </label>
            <label class="detail-field">
                <span>Дата начала</span>
                <input id="drawer-date" type="date" value="${dateInputValue(task.start_time)}">
            </label>
            <label class="detail-field">
                <span>Длительность, мин</span>
                <input id="drawer-duration" type="number" min="0" step="1" value="${durationMinutes(task)}">
            </label>
        </div>
        <section class="drawer-progress-section">
            <div class="section-heading"><h3>Прогресс</h3><span>Сохраняется в задаче</span></div>
            <div class="drawer-progress">
                <input id="drawer-progress" type="range" min="0" max="100" step="1" value="${task.progress}">
                <strong id="drawer-progress-value">${task.progress}%</strong>
            </div>
        </section>
        <div class="drawer-actions">
            <span class="save-hint">Изменения сохраняются после выбора или выхода из поля</span>
            <button id="delete-task" class="danger-button" type="button">Удалить задачу</button>
        </div>
    `;
}

function bindDrawerEvents(task) {
    document.querySelector("#drawer-title").addEventListener("change", async (event) => {
        const title = event.target.value.trim();
        if (!title) {
            event.target.value = task.title;
            showToast("Название не может быть пустым");
            return;
        }
        await updateTask(task, {title});
    });
    document.querySelector("#drawer-description").addEventListener("change", (event) => updateTask(task, {description: event.target.value.trim()}));
    document.querySelector("#drawer-column").addEventListener("change", (event) => moveTask(task, Number(event.target.value)));
    document.querySelector("#drawer-priority").addEventListener("change", (event) => updateTask(task, {priority: priorityValues[event.target.value]}));
    document.querySelector("#drawer-date").addEventListener("change", (event) => {
        const startTime = event.target.value ? new Date(`${event.target.value}T09:00:00`).toISOString() : new Date().toISOString();
        updateTask(task, {start_time: startTime});
    });
    document.querySelector("#drawer-duration").addEventListener("change", (event) => {
        const minutes = Math.max(0, Number(event.target.value) || 0);
        updateTask(task, {duration: Math.round(minutes * 60_000_000_000)});
    });
    document.querySelector("#drawer-progress").addEventListener("input", (event) => {
        document.querySelector("#drawer-progress-value").textContent = `${event.target.value}%`;
    });
    document.querySelector("#drawer-progress").addEventListener("change", (event) => updateTask(task, {progress: Number(event.target.value)}));
    document.querySelector("#drawer-done").addEventListener("change", async (event) => {
        if (event.target.checked) {
            await updateTask(task, {is_done: true, column_id: state.activeBoard.done_column_id, progress: 100});
            return;
        }
        const target = columns().find((column) => column.id !== state.activeBoard.done_column_id);
        if (!target) {
            event.target.checked = true;
            showToast("Сначала создайте рабочую колонку");
            return;
        }
        await updateTask(task, {is_done: false, column_id: target.id});
    });
    document.querySelector("#delete-task").addEventListener("click", () => {
        openConfirmDialog({
            title: "Удалить задачу?",
            message: `Задача «${task.title}» будет удалена без возможности восстановления.`,
            onConfirm: async () => {
                const deleted = await runAction(() => api(`/tasks/${task.id}`, {method: "DELETE"}), "Не удалось удалить задачу");
                if (deleted === false) return false;
                closeDrawer();
                await loadTasks();
                render();
                showToast("Задача удалена");
                return true;
            }
        });
    });
}

function taskPayload(task, changes = {}) {
    return {
        id: task.id,
        title: task.title,
        description: task.description ?? "",
        user_id: task.user_id,
        column_id: task.column_id,
        position: task.position,
        progress: task.progress,
        is_done: task.is_done,
        start_time: task.start_time,
        duration: task.duration,
        priority: task.priority,
        ...changes
    };
}

async function updateTask(task, changes, successMessage = "Задача обновлена") {
    const succeeded = await runAction(
        () => api("/tasks", {method: "PUT", body: JSON.stringify(taskPayload(task, changes))}),
        "Не удалось обновить задачу"
    );
    if (succeeded === false) return false;
    await loadTasks();
    render();
    if (activeTaskId) openDrawer(activeTaskId);
    showToast(successMessage);
    return true;
}

async function moveTask(task, columnId) {
    const target = columns().find((column) => column.id === columnId);
    if (!target) return;
    const isDone = columnId === state.activeBoard.done_column_id;
    const nextPosition = state.tasks.filter((item) => item.column_id === columnId).length;
    await updateTask(task, {
        column_id: columnId,
        position: nextPosition,
        is_done: isDone,
        progress: isDone ? 100 : task.progress
    }, `Задача перемещена в «${target.title}»`);
}

function openTaskDialog() {
    if (!state.activeBoard || !backlogColumn()) return;
    taskForm.reset();
    taskDialog.showModal();
    requestAnimationFrame(() => document.querySelector("#new-task-title").focus());
}

function closeTaskDialog() {
    taskDialog.close();
}

async function createTask() {
    const title = document.querySelector("#new-task-title").value.trim();
    if (!title) return;
    const column = backlogColumn();
    if (!column) {
        showToast("Сначала создайте рабочую колонку");
        return;
    }
    const columnId = column.id;
    const payload = {
        title,
        description: document.querySelector("#new-task-description").value.trim(),
        user_id: userId,
        column_id: columnId,
        position: state.tasks.filter((item) => item.column_id === columnId).length,
        progress: 0,
        is_done: false,
        start_time: new Date().toISOString(),
        duration: 0,
        priority: priorityValues[document.querySelector("#new-task-priority").value]
    };
    const created = await runAction(
        () => api("/tasks", {method: "POST", body: JSON.stringify(payload)}),
        "Не удалось создать задачу"
    );
    if (created === false) return;
    await loadTasks();
    render();
    closeTaskDialog();
    showToast("Задача создана");
}

function createBoard() {
    boardForm.reset();
    document.querySelector("#board-name").value = "Новая доска";
    boardDialog.showModal();
    requestAnimationFrame(() => document.querySelector("#board-name").select());
}

async function submitBoard() {
    const title = document.querySelector("#board-name").value.trim();
    if (!title) return false;
    const created = await runAction(
        () => api("/boards", {method: "POST", body: JSON.stringify({owner_user_id: userId, title})}),
        "Не удалось создать доску"
    );
    if (created === false) return false;
    boardDialog.close();
    await loadBoards(created.id, false);
    showToast("Доска создана");
    return true;
}

function createColumn() {
    if (!state.activeBoard) return;
    editingColumnId = null;
    columnForm.reset();
    document.querySelector("#column-dialog-eyebrow").textContent = "Новая колонка";
    document.querySelector("#column-dialog-title").textContent = "Добавить на доску";
    document.querySelector("#delete-column").hidden = true;
    document.querySelector("#save-column").textContent = "Создать колонку";
    columnDialog.showModal();
    requestAnimationFrame(() => document.querySelector("#column-name").focus());
}

function openColumnMenu(columnId) {
    const column = columns().find((item) => item.id === columnId);
    if (!column) return;
    editingColumnId = column.id;
    document.querySelector("#column-dialog-eyebrow").textContent = "Настройки колонки";
    document.querySelector("#column-dialog-title").textContent = column.title;
    document.querySelector("#column-name").value = column.title;
    document.querySelector("#delete-column").hidden = false;
    document.querySelector("#delete-column").disabled = column.id === state.activeBoard.done_column_id;
    document.querySelector("#save-column").textContent = "Сохранить";
    columnDialog.showModal();
    requestAnimationFrame(() => document.querySelector("#column-name").select());
}

async function submitColumn() {
    const title = document.querySelector("#column-name").value.trim();
    if (!title) return false;
    if (editingColumnId) {
        const column = columns().find((item) => item.id === editingColumnId);
        if (!column) return false;
        const updated = await updateColumn(column, {title});
        if (!updated) return false;
        columnDialog.close();
        return true;
    }

    const position = columns().reduce((maximum, column) => Math.max(maximum, column.position), -1) + 1;
    const created = await runAction(
        () => api(`/boards/${state.activeBoard.id}/columns`, {
            method: "POST",
            body: JSON.stringify({title, position})
        }),
        "Не удалось создать колонку"
    );
    if (created === false) return false;
    columnDialog.close();
    await refreshActiveBoard();
    showToast("Колонка создана");
    return true;
}

async function updateColumn(column, changes, refresh = true) {
    const result = await runAction(
        () => api(`/boards/${state.activeBoard.id}/columns/${column.id}`, {
            method: "PUT",
            body: JSON.stringify({title: column.title, position: column.position, ...changes})
        }),
        "Не удалось обновить колонку"
    );
    if (result === false) return false;
    if (refresh) await refreshActiveBoard();
    showToast("Колонка обновлена");
    return true;
}

async function reorderColumn(draggedColumnId, targetColumnId, placeAfter) {
    if (draggedColumnId === targetColumnId) return;
    const ordered = columns();
    const dragged = ordered.find((column) => column.id === draggedColumnId);
    if (!dragged) return;

    const reordered = ordered.filter((column) => column.id !== draggedColumnId);
    const targetIndex = reordered.findIndex((column) => column.id === targetColumnId);
    if (targetIndex < 0) return;
    reordered.splice(targetIndex + (placeAfter ? 1 : 0), 0, dragged);

    const changed = reordered
        .map((column, position) => ({column, position}))
        .filter(({column, position}) => column.position !== position);
    if (!changed.length) return;

    const updated = await runAction(
        () => Promise.all(changed.map(({column, position}) => api(`/boards/${state.activeBoard.id}/columns/${column.id}`, {
            method: "PUT",
            body: JSON.stringify({title: column.title, position})
        }))),
        "Не удалось изменить порядок колонок"
    );
    await refreshActiveBoard();
    if (updated !== false) showToast("Порядок колонок сохранён");
}

function requestColumnDelete(column) {
    if (column.id === state.activeBoard.done_column_id) {
        showToast("Системную колонку «Готово» удалить нельзя");
        return;
    }
    const taskCount = state.tasks.filter((task) => task.column_id === column.id).length;
    const candidates = taskCount > 0 ? columns().filter((item) => item.id !== column.id) : [];
    columnDialog.close();
    openConfirmDialog({
        title: "Удалить колонку?",
        message: taskCount > 0
            ? `В колонке «${column.title}» находится ${taskCount} задач. Перед удалением выберите, куда их перенести.`
            : `Пустая колонка «${column.title}» будет удалена.`,
        candidates,
        onConfirm: (targetColumnId) => deleteColumn(column, targetColumnId)
    });
}

async function deleteColumn(column, targetColumnId = 0) {
    const suffix = targetColumnId ? `?target_column_id=${targetColumnId}` : "";
    const deleted = await runAction(
        () => api(`/boards/${state.activeBoard.id}/columns/${column.id}${suffix}`, {method: "DELETE"}),
        "Не удалось удалить колонку"
    );
    if (deleted === false) return false;
    await refreshActiveBoard();
    showToast("Колонка удалена");
    return true;
}

function openConfirmDialog({title, message, candidates = [], onConfirm}) {
    confirmAction = onConfirm;
    document.querySelector("#confirm-title").textContent = title;
    document.querySelector("#confirm-message").textContent = message;
    const targetField = document.querySelector("#confirm-target-field");
    const target = document.querySelector("#confirm-target");
    targetField.hidden = candidates.length === 0;
    target.innerHTML = candidates.map((column) => `<option value="${column.id}">${escapeHTML(column.title)}</option>`).join("");
    confirmDialog.showModal();
}

function closeConfirmDialog() {
    confirmDialog.close();
    confirmAction = null;
}

async function submitConfirmation() {
    if (!confirmAction) return false;
    const targetColumnId = document.querySelector("#confirm-target-field").hidden
        ? 0
        : Number(document.querySelector("#confirm-target").value);
    const confirmed = await confirmAction(targetColumnId);
    if (confirmed === false) return false;
    closeConfirmDialog();
    return true;
}

async function runAction(action, errorPrefix) {
    try {
        return await action();
    } catch (error) {
        showToast(`${errorPrefix}: ${error.message}`);
        return false;
    }
}

function showToast(message) {
    clearTimeout(toastTimer);
    toast.textContent = message;
    toast.classList.add("visible");
    toastTimer = setTimeout(() => toast.classList.remove("visible"), 3200);
}

searchInput.addEventListener("input", renderBoard);
priorityFilter.addEventListener("change", renderBoard);
document.querySelector("#new-task-top").addEventListener("click", openTaskDialog);
document.querySelector("#new-board").addEventListener("click", createBoard);
document.querySelector("#new-column").addEventListener("click", createColumn);
document.querySelector("#close-dialog").addEventListener("click", closeTaskDialog);
document.querySelector("#cancel-task").addEventListener("click", closeTaskDialog);
taskForm.addEventListener("submit", async (event) => {
    event.preventDefault();
    await createTask();
});
document.querySelector("#close-board-dialog").addEventListener("click", () => boardDialog.close());
document.querySelector("#cancel-board").addEventListener("click", () => boardDialog.close());
boardForm.addEventListener("submit", async (event) => {
    event.preventDefault();
    await submitBoard();
});
document.querySelector("#close-column-dialog").addEventListener("click", () => columnDialog.close());
document.querySelector("#cancel-column").addEventListener("click", () => columnDialog.close());
columnForm.addEventListener("submit", async (event) => {
    event.preventDefault();
    await submitColumn();
});
document.querySelector("#delete-column").addEventListener("click", () => {
    const column = columns().find((item) => item.id === editingColumnId);
    if (column) requestColumnDelete(column);
});
columnDialog.addEventListener("close", () => {
    editingColumnId = null;
});
document.querySelector("#close-confirm-dialog").addEventListener("click", closeConfirmDialog);
document.querySelector("#cancel-confirm").addEventListener("click", closeConfirmDialog);
confirmForm.addEventListener("submit", async (event) => {
    event.preventDefault();
    await submitConfirmation();
});
confirmDialog.addEventListener("close", () => {
    confirmAction = null;
});
document.querySelector("#close-drawer").addEventListener("click", closeDrawer);
document.querySelector("#drawer-backdrop").addEventListener("click", closeDrawer);

document.addEventListener("keydown", (event) => {
    if ((event.metaKey || event.ctrlKey) && event.key.toLocaleLowerCase("ru") === "k") {
        event.preventDefault();
        searchInput.focus();
    }
    if (event.key === "Escape" && drawerLayer.classList.contains("open")) closeDrawer();
});

onRouteChange(async () => {
    const route = readRoute();
    if (route.userId !== userId) {
        window.location.reload();
        return;
    }
    const next = state.boards.find((item) => item.id === route.boardId);
    if (!next) {
        await runAction(() => loadBoards(route.boardId), "Не удалось открыть доску");
        return;
    }
    if (next.id === state.activeBoard?.id) return;
    state.activeBoard = next;
    localStorage.setItem(selectedBoardStorageKey, String(next.id));
    closeDrawer();
    await runAction(loadTasks, "Не удалось открыть доску");
    render();
});

render();
runAction(() => loadBoards(initialRoute.boardId), "Не удалось загрузить доски");
