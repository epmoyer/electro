var App = App || {}; // Create namespace

(function () {
    "use strict";

    App.main = () => {
        console.log("main()");

        // -----------------------
        // menu-tree: Find all items (spans) in all menu-tree(s)
        // -----------------------
        const menu_trees = document.getElementsByClassName("menu-tree");
        const all_menu_spans = [];
        for (let menu_tree of menu_trees) {
            const menu_spans = menu_tree.getElementsByTagName("span");
            // console.log(menu_spans);
            for (let span of menu_spans) {
                all_menu_spans.push(span);
            }
        }

        // -----------------------
        // menu-tree: Set item (span) click handler
        // -----------------------
        for (let span of all_menu_spans) {
            span.addEventListener("click", function () {
                // Unselect all items in all menu-tree(s)
                for (let old_span of all_menu_spans) {
                    old_span.classList.remove("selected");
                }
                // console.log("Clicked a span");
                this.classList.toggle("selected");
                // Navigate to link, if this span contains one
                const anchors = this.getElementsByTagName("a");
                if (anchors) {
                    for (let anchor of anchors) {
                        window.location.href = anchor.href;
                        // There should be only one anchor, but break anyway.
                        break;
                    }
                }
            });
        }

        // -----------------------
        // menu-tree: Select the current document (at page load)
        // -----------------------
        for (let span of all_menu_spans) {
            if (span.id == "menuitem_doc_" + App.globalConfig.currentDocumentName) {
                span.classList.add("selected");
            }
        }

        // -----------------------
        // menu-tree: Set caret (child folding) click handler
        // -----------------------
        var toggler = document.getElementsByClassName("caret");
        var i;

        for (i = 0; i < toggler.length; i++) {
            toggler[i].addEventListener("click", function (e) {
                this.closest("li").querySelector(".nested").classList.toggle("active");
                this.classList.toggle("caret-down");
                e.stopPropagation();
            });
        }

        // -----------------------
        // Search handler
        // -----------------------
        var searchBox = document.getElementById("search-text");
        console.log(searchBox);
        searchBox.addEventListener(
            "keydown",
            function (event) {
                if (event.keyCode == 13) {
                    event.preventDefault();
                    App.onSearch();
                }
            },
            false
        );
    };

    App.onSearch = () => {
        console.log("onSearch()");
        var searchText = document.getElementById('search-text').value;
        console.log('searchText', searchText);
    };
})(); // "use strict" wrapper
