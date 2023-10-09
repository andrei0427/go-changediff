(function () {
  window.changediff = function () {};

  window.changediff.init = function (options) {
    this.options = options;
    const iframe = (this.iframe = document.createElement("iframe"));
    const backdrop = (this.backdrop = document.createElement("div"));

    iframe.src = "/widget/".concat(options.appKey).concat("?embed=1");
    iframe.style.top = "0";
    iframe.style.right = "-500px";
    iframe.style.position = "absolute";
    iframe.style.width = "499px";
    iframe.style.height = window.innerHeight.toString().concat("px");
    iframe.style.border = "none";
    iframe.style.borderLeft = "1px solid #eee";
    iframe.style.transition = "right 0.2s ease-out";

    backdrop.style.opacity = 0.3;
    backdrop.style.backgroundColor = "black";

    window.addEventListener("resize", function () {
      iframe.height = window.innerHeight.toString().concat("px");
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
            iframe.style.right = "-500px";
          }
        }
      },
      false
    );

    document.body.appendChild(iframe);
  };

  window.changediff.open = function () {
    this.iframe.style.right = "0px";
  };
})();
