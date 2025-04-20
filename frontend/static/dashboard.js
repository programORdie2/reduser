let projects = [];

document.getElementById("createProjectBtn").onclick = () => {
	const name = prompt("New project name:", "Project");
	if (name) {
		createProject(name);
		load();
	}
};

function renderProjects(projects) {
	const container = document.getElementById("projectsList");
	container.innerHTML = "";
	projects.forEach((p) => {
		const div = document.createElement("div");
		div.textContent = p.name;
		div.className = "item";
		div.onclick = () => openProject(p.id);
		container.appendChild(div);
	});
}

function openProject(id) {
	document.location = `/project/${id}`;
}

async function load() {
	projects = (await loadProjects()) || [];
	renderProjects(projects);
}

load();
