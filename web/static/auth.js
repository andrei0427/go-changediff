const sb = supabase.createClient(
  "https://nplfvngvqamyeanblotv.supabase.co",
  "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJpc3MiOiJzdXBhYmFzZSIsInJlZiI6InBxYmltaGtra2ZrZGdkbXZ4d2x3Iiwicm9sZSI6ImFub24iLCJpYXQiOjE2OTIwNDM2NjcsImV4cCI6MjAwNzYxOTY2N30.OcQzGB79P63ugphDGdh3Amc6OtNtTpH0f6JIZXLDqqw",
  {
    auth: {
      authRefreshToken: true,
    },
  }
);

document.querySelectorAll(".btnLogin").forEach((element) => {
  element.addEventListener("click", handleLoginRegisterClick);
});

const AUTH_USER_COOKIE_NAME = "authUser";
function isUserAuthenticated() {
  const authUser = document.cookie
    .split("; ")
    .find((row) => row.startsWith(AUTH_USER_COOKIE_NAME))
    ?.split("=")[1];
  if (authUser) {
    handleUserAuthenticated();
    return;
  }

  sb.auth.getSession().then(({ data, error }) => {
    if (data.session != null) {
      handleUserAuthenticated(data.session.access_token);
    } else if (window.location.pathname === "/dashboard") {
      window.location = "/";
    }
  });
}

function handleUserAuthenticated(access_token) {
  document.querySelectorAll(".btnLogin").forEach((ele) => {
    ele.innerHTML = "Dashboard";
    ele.addEventListener("click", () => {
      window.location = "/dashboard";
    });
  });

  if (access_token) {
    document.cookie = `${AUTH_USER_COOKIE_NAME}=${access_token};max-age=3600;secure`;
    window.location = "/dashboard";
  }

  document.getElementById("signInCTA")?.style.setProperty("display", "none");
}

async function handleLoginRegisterClick() {
  const { data, error } = await sb.auth.signInWithOAuth({
    provider: "google",
  });

  if (error) {
    throw new Error(error);
  }

  if (data.session != null) {
    handleUserAuthenticated(data.session.access_token);
  }
}

async function handleLogout() {
  document.cookie = AUTH_USER_COOKIE_NAME + "=; Max-Age=-99999999;";

  const { error } = await sb.auth.signOut();

  if (error) {
    console.error(error);
  }

  window.location = "/";
}

isUserAuthenticated();
