const apiBase = "http://localhost:8080/api";
let jwt = "";

if (document.cookie) {
	const cookies = document.cookie.split(";");
	for (const cookie of cookies) {
		const [name, value] = cookie.split("=");
		if (name === "jwt") {
			jwt = value;
		}
	}
}

async function login(username, password) {
	const res = await fetch(`${apiBase}/login`, {
		method: "POST",
		body: JSON.stringify({ username, password }),
		headers: { "Content-Type": "application/json" },
	});
	const data = await res.json();
	if (data.token) {
		jwt = data.token;
		document.cookie = `jwt=${jwt};path=/;max-age=${60 * 60 * 24 * 30}`;
		document.location = "/dashboard";
	} else alert("Login failed");
}

async function register(username, password) {
	await fetch(`${apiBase}/register`, {
		method: "POST",
		body: JSON.stringify({ username, password }),
		headers: { "Content-Type": "application/json" },
	});
	login(username, password);
}

async function loadProjects() {
	const res = await fetch(`${apiBase}/projects`, {
		headers: { Authorization: "Bearer " + jwt },
	});
	const list = await res.json();
	return list;
}

async function createProject(name) {
	await fetch(`${apiBase}/projects`, {
		method: "POST",
		headers: {
			"Content-Type": "application/json",
			Authorization: "Bearer " + jwt,
		},
		body: JSON.stringify({ name }),
	});
}

async function loadProject(currentProjectId) {
	const res = await fetch(`${apiBase}/projects/${currentProjectId}`, {
		headers: { Authorization: "Bearer " + jwt },
	});
	const project = await res.json();
	return project;
}

async function createTable(currentProjectId, name) {
	await fetch(`${apiBase}/projects/${currentProjectId}/tables`, {
		method: "POST",
		headers: {
			"Content-Type": "application/json",
			Authorization: "Bearer " + jwt,
		},
		body: JSON.stringify({ name }),
	});
}

async function newVar(currentProjectId, currentTableId, name, type) {
	await fetch(
		`${apiBase}/projects/${currentProjectId}/tables/${currentTableId}/variables`,
		{
			method: "POST",
			headers: {
				"Content-Type": "application/json",
				Authorization: "Bearer " + jwt,
			},
			body: JSON.stringify({ name, type, value: "" }),
		},
	);
}

async function setVariable(currentProjectId, currentTableId, type, name) {
	await fetch(
		`${apiBase}/projects/${currentProjectId}/tables/${currentTableId}/variables/${name}`,
		{
			method: "PUT",
			headers: {
				"Content-Type": "application/json",
				Authorization: "Bearer " + jwt,
			},
			body: JSON.stringify({ new_type: type }),
		},
	);
}

async function deleteVariable(name, currentProjectId, currentTableId) {
	await fetch(
		`${apiBase}/projects/${currentProjectId}/tables/${currentTableId}/variables/${name}`,
		{
			method: "DELETE",
			headers: { Authorization: "Bearer " + jwt },
		},
	);
}

async function updateProject(currentProjectId, name) {
	await fetch(`${apiBase}/projects/${currentProjectId}`, {
		method: "PUT",
		headers: {
			"Content-Type": "application/json",
			Authorization: "Bearer " + jwt,
		},
		body: JSON.stringify({ name }),
	});
}

async function updateTableName(currentProjectId, currentTableId, name) {
	await fetch(
		`${apiBase}/projects/${currentProjectId}/tables/${currentTableId}`,
		{
			method: "PUT",
			headers: {
				"Content-Type": "application/json",
				Authorization: "Bearer " + jwt,
			},
			body: JSON.stringify({ name }),
		},
	);
}

async function deleteProject(currentProjectId) {
	await fetch(`${apiBase}/projects/${currentProjectId}`, {
		method: "DELETE",
		headers: { Authorization: "Bearer " + jwt },
	});
}

async function deleteTable(currentProjectId, currentTableId) {
	await fetch(
		`${apiBase}/projects/${currentProjectId}/tables/${currentTableId}`,
		{
			method: "DELETE",
			headers: { Authorization: "Bearer " + jwt },
		},
	);
}

document.getElementById("logoutBtn").onclick = () => {
	document.cookie = "jwt=;path=/;max-age=0";
	location = "/";
};
