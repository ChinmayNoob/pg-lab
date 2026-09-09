/* pg-lab reusable quiz component.
   Markup contract:
   <div class="quiz">
     <p class="question">…</p>
     <div class="options">
       <button data-correct="false">…</button>
       <button data-correct="true">…</button>
     </div>
     <p class="explanation">…</p>
   </div>
   Options are shuffled at load so correctness never leaks via position. */
(function () {
  "use strict";

  function shuffle(list) {
    var out = Array.prototype.slice.call(list);
    for (var i = out.length - 1; i > 0; i--) {
      var j = Math.floor(Math.random() * (i + 1));
      var tmp = out[i]; out[i] = out[j]; out[j] = tmp;
    }
    return out;
  }

  function initQuiz(quiz) {
    var options = quiz.querySelector(".options");
    if (!options) return;

    shuffle(options.querySelectorAll("button")).forEach(function (btn) {
      options.appendChild(btn);
      btn.addEventListener("click", function () {
        if (quiz.classList.contains("answered")) return;
        quiz.classList.add("answered");
        btn.classList.add(btn.dataset.correct === "true" ? "correct" : "incorrect");
      });
    });
  }

  function initAll(root) {
    (root || document).querySelectorAll(".quiz").forEach(initQuiz);
  }

  if (document.readyState === "loading") {
    document.addEventListener("DOMContentLoaded", function () { initAll(); });
  } else {
    initAll();
  }

  window.pglabQuizzes = { init: initAll };
})();
