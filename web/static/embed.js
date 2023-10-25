(function () {
  window.changediff = function () {};

  window.changediff.init = function (options = {}) {
    this.options = options;
    const iframe = (this.iframe = document.createElement("iframe"));
    const backdrop = (this.backdrop = document.createElement("div"));

    iframe.src = "/widget/".concat(options.appKey).concat("?embed=1");
    iframe.style.top = "0";
    iframe.style.right = "-500px";
    iframe.style.position = "absolute";
    iframe.style.width = "499px";
    iframe.style.border = "none";
    iframe.style.borderLeft = "1px solid #eee";
    iframe.style.transition = "right 0.2s ease-out";
    iframe.style.zIndex = 999;
    iframe.style.backgroundColor = "white";

    backdrop.style.opacity = 0;
    backdrop.style.backgroundColor = "black";
    backdrop.style.zIndex = 998;
    backdrop.style.position = "absolute";
    backdrop.style.top = "0";
    backdrop.style.left = "0";
    backdrop.style.width = "100%";
    backdrop.style.transition = "opacity 0.2s";
    backdrop.style.display = "none";

    backdrop.addEventListener("click", function () {
      window.changediff.close();
    });

    iframe.style.height = backdrop.style.height = window.innerHeight
      .toString()
      .concat("px");

    window.addEventListener("resize", function () {
      iframe.height = backdrop.style.height = window.innerHeight
        .toString()
        .concat("px");
    });

    window.addEventListener(
      "message",
      function (event) {
        if (
          !["localhost", "127.0.0.1", "https://changediff.io"].some((o) =>
            event.origin.includes(o)
          )
        ) {
          return;
        }

        switch (event.data) {
          case "close": {
            this.changediff.close();
            break;
          }
          case "loaded": {
            const { id, name, email, role, info } = options?.userInfo;
            iframe.contentWindow.postMessage(
              {
                type: "userInfo",
                data: JSON.stringify({
                  id,
                  name,
                  email,
                  role,
                  info,
                }),
              },
              // FIXME
              "http://localhost:8000"
            );

            break;
          }
        }
      },
      false
    );

    document.body.appendChild(iframe);
    document.body.appendChild(backdrop);
  };

  window.changediff.close = function () {
    this.iframe.style.right = "-500px";
    this.backdrop.style.opacity = 0;

    setTimeout(() => {
      this.backdrop.style.display = "none";
    }, 300);
  };

  window.changediff.open = function () {
    this.iframe.style.right = "0px";
    this.backdrop.style.opacity = 0.3;
    this.backdrop.style.display = "block";
  };
})();
