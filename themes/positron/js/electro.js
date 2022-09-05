var App = App || {}; // Create namespace

(function () {
    "use strict";

    App.state = {
        allMenuSpans: null
    };

    App.main = () => {
        // -----------------------
        // menu-tree: Find all items (spans) in all menu-tree(s)
        // -----------------------
        const menu_trees = document.getElementsByClassName("menu-tree");
        App.state.allMenuSpans = [];
        for (let menu_tree of menu_trees) {
            const menu_spans = menu_tree.getElementsByTagName("span");
            for (let span of menu_spans) {
                App.state.allMenuSpans.push(span);
            }
        }

        // -----------------------
        // menu-tree: Set item (span) click handler
        // -----------------------
        for (let span of App.state.allMenuSpans) {
            span.addEventListener("click", function () {
                App.onClickMenuItem(this);
                // // Unselect all items in all menu-tree(s)
                // for (let old_span of App.state.allMenuSpans) {
                //     old_span.classList.remove("selected");
                //     old_span.classList.remove("navigating");
                // }

                // const pageId = this.dataset.documentName;
                // const targetHeadingId = this.dataset.targetHeadingId;
                
                // // -----------------------------
                // // Make the target page visible
                // // -----------------------------
                // App.showPage(pageId);

                // // ----------------------------
                // // Scroll to the target heading (or to top if no HeadingId exists)
                // // -----------------------------
                // App.scrollToHash(targetHeadingId);

                // // -----------------------------
                // // Select the clicked menu item
                // // -----------------------------
                // this.classList.add("selected");

                // // If in responsive narrow-screen, re-hide the menu.
                // document.getElementById("sidebar-container").classList.remove("force-show");
            });
        }

        // -----------------------
        // menu-tree: Select the current document (at page load)
        // -----------------------
        for (let span of App.state.allMenuSpans) {
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
        // console.log(searchBox);
        searchBox.addEventListener(
            "keydown",
            function (event) {
                if (event.keyCode == 13) {
                    event.preventDefault();
                    if (App.globalConfig.singleFile){
                        const searchText = document.getElementById("search-text").value;
                        App.doSearch(searchText);
                    }
                    else{
                        App.onSearch();
                    }
                }
            },
            false
        );

        // -----------------------
        // Home button
        // -----------------------
        if(App.globalConfig.singleFile){
            // There are 2 versions of the home button. The default is an anchor tag
            // to index.html, for static web sites.  For single file docs we need to hide that
            // one and enable the one which contains a pageId link parameter (to the doc_info page).
            document.getElementById("home-button-static").style.display = 'none';
            document.getElementById("home-button-single-file").style.display = 'inline';
        }

        // -----------------------
        // Get search data
        // -----------------------
        const config = App.globalConfig;
        config.searchIndex = lunr(function () {
            this.field("title");
            this.field("text");
            this.field("heading");
            this.ref("location");

            for (var i = 0; i < App.searchData.docs.length; i++) {
                var doc = App.searchData.docs[i];
                this.add(doc);
                config.documents[doc.location] = doc;
            }
        });

        // Execute search query (if requested in url)
        const searchText = App.getUrlValue('query');
        if(searchText){
            document.getElementById("search-text").value = searchText;
            console.log('query', searchText);
            App.doSearch(searchText);
        }
        
        // Navigate to search result location (if requested in url)
        const pageId = App.getUrlValue('pageId');
        const headingId = App.getUrlValue('headingId');
        if(pageId){
            console.log('Going to location: pageId:' + pageId + ' headingId:' + headingId);
            App.showPage(pageId);
            if(headingId){
                App.scrollToHash(headingId);
            }
        }

        const hideWatermark = App.getUrlValue('nowm');
        if (hideWatermark == 'true'){
            console.log('Hiding watermark because nowm set in URL.');
            document.getElementById("watermark-text").innerHTML = "";
        }

        // -----------------------
        //  Set handler
        // -----------------------
        const main_element = document.getElementsByClassName("main-container")[0];
        main_element.addEventListener('touchmove', App.onMainTouchMove, false);

        
        // Show version
        console.log('Built with Electro ' + App.globalConfig.electroVersion);
    };

    App.onClickMenuItem = (self) => {
        console.log("App.onClickMenuItem");
        // Unselect all items in all menu-tree(s)
        for (let old_span of App.state.allMenuSpans) {
            old_span.classList.remove("selected");
            old_span.classList.remove("navigating");
        }
        
        if (App.globalConfig.singleFile){
            App.onClickMenuSingleFile(self);
        }
        else{
            App.onClickMenuStaticSite(self);
        }
    };
    
    App.onClickMenuSingleFile = (self) => {
        console.log("App.onClickMenuSingleFile");
        const pageId = self.dataset.documentName;
        const targetHeadingId = self.dataset.targetHeadingId;
        
        // -----------------------------
        // Make the target page visible
        // -----------------------------
        App.showPage(pageId);

        // ----------------------------
        // Scroll to the target heading (or to top if no HeadingId exists)
        // -----------------------------
        App.scrollToHash(targetHeadingId);

        // -----------------------------
        // Select the clicked menu item
        // -----------------------------
        self.classList.add("selected");

        // If in responsive narrow-screen, re-hide the menu.
        document.getElementById("sidebar-container").classList.remove("force-show");
    };

    App.onClickMenuStaticSite = (self) => {
        console.log("App.onClickMenuStaticSite");
        // Navigate to link, if this span contains one
        var target_url = null;
        const anchors = self.getElementsByTagName("a");
        if (anchors) {
            for (let anchor of anchors) {
                target_url = anchor.href;
                // There should be only one anchor, but break anyway.
                break;
            }
        }
        if (target_url !== null){
            // if(App.globalConfig.currentDocumentName in 
            if (target_url.indexOf(App.globalConfig.currentDocumentName) > -1) {
                // We are navigating to the current page
                self.classList.add("selected");
            }
            else{
                // We are navigating away from the current page
                self.classList.add("navigating");
            }
            window.location.href = target_url;
        }
        else {
            self.classList.add("selected");
        }
    };

    App.showPage = (pageId) => {
        const pages = document.getElementsByClassName('content-page');
        for (const page of pages) {
            if(page.id == pageId){
                page.style.display = 'inline';
            } else {
                page.style.display = 'none';
            }
        }
    };

    App.scrollToHash = (hashName) => {
        if(!hashName){
            // Scroll to the top if no hashName was given
            var content_div = document.getElementsByClassName("main-container")[0];
            content_div.scrollTop = 0;
            return;
        }
        const target = "#" + hashName;
        if(location.hash == target){
            // Browser thinks we are already at the target, but we might have scrolled away
            // from it, so blank it and then set it back to force the document to scroll
            location.hash = "";
            location.hash = "#" + hashName;
        }
        else {
            location.hash = "#" + hashName;
        }
    };

    App.onMainTouchMove = (e) => {
        if(document.getElementById("sidebar-container").classList.contains("force-show")){
            // Don't allow the main window to be scrolled while the sidebar is 
            // deployed in "mobile" (narrow screen) mode.
            e.preventDefault();
        }
    };

    App.toggleSidebar = () => {
        console.log('toggleSidebar()');
        document.getElementById("sidebar-container").classList.toggle("force-show");
    };

    App.getUrlValue  = (VarSearch) => {
        var SearchString = window.location.search.substring(1);
        var VariableArray = SearchString.split('&');
        for(var i = 0; i < VariableArray.length; i++){
            var KeyValuePair = VariableArray[i].split('=');
            if(KeyValuePair[0] == VarSearch){
                return decodeURI(KeyValuePair[1]);
            }
        }
    };

    App.onSearch = () => {
        // Executed on static sites when user enters search text and presses return.
        //
        // The search is encoded as a query parameter on the URL, and the search page
        // (search.html) is loaded. The actual searching and displaying of results will
        // happen on that page.
        console.log("onSearch()");
        var searchText = document.getElementById("search-text").value;
        console.log("searchText", searchText);
        window.location.href = 'search.html?query=' + searchText;
    };

    App.doSearch = (searchText) => {
        // console.log("doSearch()");
        
        const results = App.search(searchText);
        // console.log(results);

        var html = '<h1>Search Results</h1>\n';
        results.forEach(result => {
            // html += '<hr>';
            if (App.globalConfig.singleFile){
                // Single file compatible link
                const location = result.location.replace(".html", "");
                const pieces = location.split("#");
                const pageId = pieces[0];
                const headingId = pieces[1];
                html += '<h3><a href="?pageId=' + pageId + '&headingId=' + headingId + '">' + result.title + '</a></h3>';
            } else {
                // Static site compatible link
                html += '<h3><a href="' + result.location + '">' + result.title + '</a></h3>';
            }
            if(result.heading){
                html += '<h4>' + result.heading + '</h4>';
            }

            // Highlight search hits
            const summary = App.highlight(searchText, result.summary);

            html += '<p>' + summary + '</p>';
        });
        if (results.length == 0){
            html += '(no results to show)';
        }
        if(!App.globalConfig.singleFile){
            // Static web page. Show search results in content area
            document.getElementById("content").innerHTML = html;
        } else {
            // Single file document. Show search results in search page and hide other pages.
            document.getElementById("search").innerHTML = html;

            // Make the search page visible
            App.showPage('search');
        }
    };

    App.highlight = (searchText, html) => {
        // Case insensitive highlight searchText in html.
        // Wraps hits with '<span class="highlight">' tag.
        const regex = new RegExp(searchText, "ig");
        const matches = html.matchAll(regex);
        const match_words = [];
        var match;
        for(match of matches){
            match_words.push(match[0]);
        }
        const unique = Array.from(new Set(match_words));
        for(match of unique){
            html = html.replace(
                match, '<span class="highlight">' + match + '</span>');
        }
        return(html);
    };

    App.search = (query) => {
        const config = App.globalConfig;
        if (config.searchIndex === null) {
            console.error("Assets for search still loading");
            return;
        }

        var resultDocuments = [];
        var results = config.searchIndex.search(query);
        for (var i = 0; i < results.length; i++) {
            var result = results[i];
            var doc = config.documents[result.ref];
            doc.summary = doc.text.substring(0, 200);
            resultDocuments.push(doc);
        }
        return resultDocuments;
    };
})(); // "use strict" wrapper
