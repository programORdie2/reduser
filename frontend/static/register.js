const loginForm = document.getElementById("loginForm");

loginForm.onsubmit = async (e) => {
	e.preventDefault();
	const username = document.getElementById("username").value;
	const password = document.getElementById("password").value;

	if (!username || !password) {
		alert("Please enter username and password");
		return;
	}

	register(username, password);
};
