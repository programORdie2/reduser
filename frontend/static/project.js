let id = "";

// Parse the url parts to get the id
const urlParts = window.location.pathname.split("/").filter(Boolean);
id = urlParts[urlParts.length - 1];

document.getElementById("delBtn").onclick = async () => {
	if (!confirm("Are you sure you want to delete this project?")) return;

	await deleteProject(id);
	window.location.href = "/dashboard";
};

document.getElementById("renameBtn").onclick = () => {
	const name = prompt("New project name:", "Project");
	if (name) {
		updateProject(id, name);
		load();
	}
};

function renderProject(project) {
	document.getElementById("projectTitle").textContent = project.name;
	const tables = project.tables || [];
	renderTables(tables);
}

function renderTables(tables) {
	const container = document.getElementById("tablesList");
	container.innerHTML = "";

	const newBtn = document.createElement("button");
	newBtn.textContent = "New Table";
	newBtn.onclick = async () => {
		const name = prompt("New table name:", "Table");
		if (name) {
			await createTable(id, name);
			load();
		}
	};
	container.appendChild(newBtn);

	tables.forEach((t) => {
		const div = document.createElement("div");
		const title = document.createElement("h3");
		title.textContent = t.name;
		div.appendChild(title);
		div.className = "item";

		const updateBtn = document.createElement("button");
		updateBtn.textContent = "Rename Table";
		updateBtn.onclick = () => {
			const name = prompt("New table name:", "Table");
			if (name) {
				updateTableName(id, t.id, name);
				load();
			}
		};
		div.appendChild(updateBtn);

		const delBtn = document.createElement("button");
		delBtn.textContent = "Delete Table";
		delBtn.onclick = () => {
			if (confirm("Are you sure you want to delete this table?")) {
				deleteTable(id, t.id);
				load();
			}
		};
		div.appendChild(delBtn);

		div.appendChild(renderVars(t.variables || [], t));

		const newVarBtn = document.createElement("button");
		newVarBtn.textContent = "New Variable";
		newVarBtn.onclick = async () => {
			const name = prompt("New variable name:", "Variable");
			if (name) {
				await newVar(id, t.id, name, "string");
				load();
			}
		};
		div.appendChild(newVarBtn);

		container.appendChild(div);
	});
}

function renderVars(vars, t) {
	const table = document.createElement("table");
	const thead = document.createElement("thead");
	const tr = document.createElement("tr");
	const th1 = document.createElement("th");
	th1.textContent = "Name";
	tr.appendChild(th1);
	const th2 = document.createElement("th");
	th2.textContent = "Type";
	tr.appendChild(th2);
	const th3 = document.createElement("th");
	th3.textContent = "Value";
	tr.appendChild(th3);
	const th4 = document.createElement("th");
	th4.textContent = "Actions";
	tr.appendChild(th4);
	thead.appendChild(tr);
	table.appendChild(thead);

	const tbody = document.createElement("tbody");
	vars.forEach((v) => {
		const tr = document.createElement("tr");
		const nameTd = document.createElement("td");
		nameTd.textContent = v.name;
		tr.appendChild(nameTd);
		const typeTd = document.createElement("td");
		typeTd.textContent = v.type;
		tr.appendChild(typeTd);
		const valueTd = document.createElement("td");
		valueTd.textContent = v.value;
		tr.appendChild(valueTd);
		const actionsTd = document.createElement("td");
		const deleteBtn = document.createElement("button");
		deleteBtn.textContent = "Delete";
		deleteBtn.onclick = async () => {
			if (confirm("Are you sure you want to delete this variable?")) {
				await deleteVariable(v.name, id, t.id);
				load();
			}
		};
		actionsTd.appendChild(deleteBtn);
		const updateBtn = document.createElement("button");
		updateBtn.textContent = "Update";
		updateBtn.onclick = async () => {
			// Replace the variable name with an input field
			const input = document.createElement("input");
			input.value = v.name;
			nameTd.textContent = "";
			nameTd.appendChild(input);

			updateBtn.hidden = true;

			// Replace the variable type with a select field
			const select = document.createElement("select");
			select.innerHTML = `<option value="string">string</option>
                                <option value="int">int</option>
                                <option value="float">float</option>
                                <option value="bool">boolean</option>`;
			typeTd.textContent = "";
			typeTd.appendChild(select);

			// Add a button to save the changes
			const saveBtn = document.createElement("button");
			saveBtn.textContent = "Save";
			saveBtn.onclick = async () => {
				const name = input.value;
				const type = select.value;
				await setVariable(id, t.id, name, type, v.name);
				load();
			};
			actionsTd.appendChild(saveBtn);
		};
		actionsTd.appendChild(updateBtn);
		tr.appendChild(actionsTd);

		tbody.appendChild(tr);
	});
	table.appendChild(tbody);
	return table;
}

async function load() {
	const project = await loadProject(id);
	renderProject(project);
}

load();
