/*
Copyright 2026 The Kubernetes Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

document.addEventListener("DOMContentLoaded", () => {
  const chapters = document.querySelector("#mdbook-sidebar .chapter");
  if (!chapters) return;
  const active = chapters.querySelector("a.active");
  const activeTop = active?.getBoundingClientRect().top;

  // mdBook's fold setting applies to chapters, but not to part headings.
  // Native details keep the categories collapsible with keyboard support.
  // Remove the empty chapter items mdBook inserts before part headings.
  chapters.querySelectorAll(":scope > .chapter-item:empty").forEach(item => item.remove());
  for (const title of chapters.querySelectorAll(":scope > .part-title")) {
    const item = document.createElement("li");
    item.className = "sidebar-section";
    const details = document.createElement("details");
    const summary = document.createElement("summary");
    summary.textContent = title.textContent;
    const section = document.createElement("ol");
    section.className = "section";

    while (title.nextElementSibling && !title.nextElementSibling.classList.contains("part-title")) {
      section.append(title.nextElementSibling);
    }
    details.open = section.querySelector("a.active") !== null;
    details.append(summary, section);
    item.append(details);
    title.replaceWith(item);
  }

  // mdBook's chapter arrows are anchors without hrefs, so Enter and Space do
  // not activate them. Use buttons with the chapter name and expanded state.
  for (const toggle of chapters.querySelectorAll(".chapter-fold-toggle")) {
    const item = toggle.closest("li");
    const button = document.createElement("button");
    button.type = "button";
    button.className = "chapter-fold-toggle";
    button.setAttribute("aria-label", `Toggle ${toggle.previousElementSibling.textContent}`);
    const updateExpanded = () => {
      button.setAttribute("aria-expanded", item.classList.contains("expanded"));
    };
    updateExpanded();
    button.addEventListener("click", () => {
      item.classList.toggle("expanded");
    });
    // Scrolling also expands mdBook's nested page headings.
    new MutationObserver(updateExpanded).observe(item, {
      attributes: true, attributeFilter: ["class"],
    });
    toggle.replaceWith(button);
  }

  // mdBook restores scrolling before the categories are grouped. Preserve the
  // selected link's position after collapsing them, without scrolling the page.
  if (active) {
    chapters.closest(".sidebar-scrollbox").scrollTop += active.getBoundingClientRect().top - activeTop;
  }
});
