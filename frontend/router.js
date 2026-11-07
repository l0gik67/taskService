export function readRoute() {
    const query = new URLSearchParams(window.location.search);
    const requestedUserId = Number(query.get("user_id"));
    const userId = Number.isInteger(requestedUserId) && requestedUserId > 0 ? requestedUserId : 1;
    const match = window.location.pathname.match(/^\/boards\/(\d+)\/?$/);
    const requestedBoardId = match ? Number(match[1]) : 0;
    const boardId = Number.isInteger(requestedBoardId) && requestedBoardId > 0 ? requestedBoardId : null;
    return {userId, boardId};
}

export function boardURL(userId, boardId = null) {
    const path = boardId ? `/boards/${boardId}` : "/boards";
    const query = new URLSearchParams({user_id: String(userId)});
    return `${path}?${query}`;
}

export function setBoardRoute(userId, boardId, replace = false) {
    const url = boardURL(userId, boardId);
    if (`${window.location.pathname}${window.location.search}` === url) return;
    window.history[replace ? "replaceState" : "pushState"]({boardId}, "", url);
}

export function onRouteChange(handler) {
    window.addEventListener("popstate", handler);
}
