"use strict";

document.addEventListener('DOMContentLoaded', function() {
    /* ===== ELEMENTS ===== */
    const displayBtn = this.getElementById('display');
    const displayLight = this.getElementById('display-light');
    const displayDark = this.getElementById('display-dark');
    const infoBtn = this.getElementById('info');
    const infoPopup = this.getElementById('info-popup');
    const infoCloseBtn = this.getElementById('info-popup-close');
    const form = this.getElementById('form');
    const logoWrapper = this.getElementById('logo-wrapper');
    const logo = this.getElementById('logo');
    const darkLogo = this.getElementById('logo-dark');
    const themeInput = this.getElementById('theme');
    const imageInput = this.getElementById('image');
    const imageOptions = this.getElementById('options');
    const customSize = this.getElementById('custom-size');
    const widthInput = this.getElementById('width');
    const heightInput = this.getElementById('height');
    const exposure = this.getElementById('exposure');
    const exposureValue = this.getElementById('exposure-value');
    const uploadBtn = this.getElementById('upload');
    const error = this.getElementById('error');
    const imagePlaceholder = this.getElementById('img-placeholder');
    const thumbnailWrapper = this.getElementById('thumbnail-wrapper');
    const thumbnail = this.getElementById('thumbnail');
    const thumbnailName = this.getElementById('thumbnail-name');
    const sizeContainer = this.getElementById('size');
    const output = this.getElementById('output');
    const copySuccess = this.getElementById('output-copy-success');
    const copyError = this.getElementById('output-copy-failure');
    const outputContainer = this.getElementById('output-wrapper');
    const submitBtn = this.getElementById('submit');
    const submitBtnText = this.getElementById('submit-btn-text');
    const submitBtnSpinner = this.getElementById('submit-btn-spinner');
    const maintainAspectRatio = this.getElementById('aspect-ratio');
    const sizeWarning = this.getElementById('size-warning');
    const sizeWarningText = this.getElementById('size-warning-text');
    const sizeRadios = sizeContainer.querySelectorAll('input[name="size"]');
    const defaultSizeRadio = this.getElementById(Array.from(sizeRadios).find(radio => radio.checked).id);
    const sizeRadioLabels = sizeContainer.getElementsByTagName('label'); 
    const widthAndHeightInputs = customSize.getElementsByTagName('input');

    /* ===== VARIABLES ===== */
    const MAX_LENGTH = 500;
    const THEME = "theme";
    const LIGHT_THEME = "light";
    const DARK_THEME = "dark";
    const FLOAT_IN_ANIMATION = 'animate-floatin';
    const size = {
        twitch: {
            width: 30,
            height: undefined,
            maxHeight: 15
        },
        discord: {
            width: 32,
            height: undefined,
            maxHeight: 62
        },
        small: {
            width: 30,
            height: undefined,
            maxHeight: MAX_LENGTH
        },
        medium: {
            width: 60,
            height: undefined,
            maxHeight: MAX_LENGTH
        },
        large: {
            width: 120,
            height: undefined,
            maxHeight: MAX_LENGTH
        }
    };

    let clipboardModalTimeout;
    let image;

    /* ===== FUNCTIONS ===== */

    /**
     * Loads page theme - defaults to localStorage, but falls back to OS settings
     */
    function onLoad() {
        let themePreference = localStorage.getItem(THEME);
        if (!themePreference) {
            themePreference = window.matchMedia('(prefers-color-scheme: dark)').matches ? DARK_THEME : LIGHT_THEME;
        }

        if (themePreference === DARK_THEME) {
            displayBtn.click();
        }
    }

    /**
     * Makes an element visible
     * 
     * @param {HTMLElement} element
     */
    function show(element) {
        element.classList.remove('sr-only');
    }

    /**
     * Hides an element
     * 
     * @param {HTMLElement} element
     */
    function hide(element) {
        element.classList.add('sr-only');
    }

    /**
     * Renders an error message to the user
     * 
     * @param {string} message - An error message.
     */
    function addErrorMessage(message) {
        show(error);
        error.textContent = message;
    };

    /**
     * Removes error message
     */
    function removeErrorMessage() {
        hide(error);
        error.textContent = '';
    }

    /**
     * Hides user options in the event of an error, which is described by `message`
     * 
     * @param {string} message - An error message.
     */
    function hideOptions(message) {
        thumbnail.src = '';
        thumbnail.alt = '';
        thumbnailName.textContent = '';
        hide(thumbnailWrapper);
        show(imagePlaceholder);

        imageInput.value = '';
        hide(imageOptions);
        imageOptions.classList.remove(FLOAT_IN_ANIMATION);
        defaultSizeRadio.checked = true;

        if (message) {
            addErrorMessage(message);
        }
    }

    /**
     * Determines whether or not the "Maintain Aspect Ratio" is checked
     * 
     * @returns {bool}
     */
    function isAspectRatioMaintained() {
        return maintainAspectRatio.checked;
    }

    /**
     * Update the size warning, based on `length`, `maxLength`, and state of `maintainAspectRatio` checkbox.
     * 
     * @param {number} length 
     * @param {number} maxLength
     * @param {isHeight} bool 
     * @returns {number}
     */
    function handleSizeWarning(length, maxLength, isHeight = true) {
        if (length <= maxLength) {
            sizeWarningText.textContent = '';
            hide(sizeWarning);
            return length;
        }

        if (isAspectRatioMaintained()) {
            sizeWarningText.textContent = `Aspect Ratio cannot be maintained. ${isHeight ? "Height" : "Width"} cannot exceed ${maxLength}.`;
            show(sizeWarning);
        }

        return maxLength;
    }

    /**
     * Update the width and height.
     * 
     * @param {number} width 
     * @param {number} height 
     * @param {number} maxHeight 
     */
    function updateWidthAndHeight(width, height, maxHeight) {
        widthInput.value = width;
        heightInput.value = handleSizeWarning(height, maxHeight);
    };

    /**
     * Function that gets the checked size type radio button.
     * 
     * @returns {string}
     */
    function getCurrentType() {
        return sizeContainer.querySelector('input:checked').value;
    }

    /**
     * Render options to user
     * 
     * @param {Image} image - The user-uploade image.
     */
    function displayOptions(image) {
        thumbnail.src = image.src;
        thumbnail.alt = image.name;
        thumbnailName.textContent = image.name;
        show(thumbnailWrapper);
        hide(imagePlaceholder);

        show(imageOptions);
        hide(error);
        error.textContent = '';
        imageOptions.classList.add(FLOAT_IN_ANIMATION);

        const type = getCurrentType();
        let width, height, maxHeight;

        // special case - if type is custom and we are NOT maintaining aspect ratio, we do not change width + height
        if (type === "custom" && !isAspectRatioMaintained()) {
            return;
        }

        if (type === "custom") {
            width = Math.min(image.width, MAX_LENGTH);
            height = getCalculatedHeight(image.width); 
            maxHeight = MAX_LENGTH;
        } else {
            width = size[type].width;
            height = size[type].height;
            maxHeight = size[type].maxHeight;
        }
        
        updateWidthAndHeight(width, height, maxHeight);
    };

    /**
     * Given an ascii height, determine width of ascii that maintains aspect ratio.
     * 
     * @param {height} height - The height of the ascii 
     * @returns {number} The new width.
     */
    function getCalculatedWidth(height) {
        return 2 * Math.round((height * image.width) / image.height);
    }

    /**
     * Given an ascii width, determine height of ascii that maintains aspect ratio.
     * 
     * @param {number} width - The width of the ascii.
     * @returns {number} The new height.
     */
    function getCalculatedHeight(width) {
        return Math.max(1, Math.round((width * image.height) / image.width / 2));
    }

    /**
     * Validates image, and builds it out.
     * 
     * @param {FileList} files The file(s) uploaded by the user. Only first image is handled.
     */
    function handleNewImage(files) {
        const img = files[0];
        const validTypes = ['image/jpeg', 'image/png'];

        if (validTypes.includes(img.type)) {
            image = new Image();
            image.src = URL.createObjectURL(img);
            image.onload = function() {
                this.setAttribute('name', img.name);
                Object.keys(size).forEach(type => size[type].height = getCalculatedHeight(size[type].width));
                displayOptions(this);
            }
        } else {
            hideOptions('File type not supported. Please upload a JPEG or PNG file.');
        }
    };

    /**
     * Toggle form in "submitting" state
     * 
     * @param {boolean} isSubmitting Flag that controls whether or not we are in "submitting" state.
     */
    function setSubmitting(isSubmitting) {
        const submittingClasses = ['cursor-not-allowed', 'bg-blue-500/90'];
        const normalClasses = ['hover:bg-blue-500/90'];

        if (isSubmitting) {
            hide(submitBtnText);
            show(submitBtnSpinner);
            submitBtn.disabled = true;
            submitBtn.classList.add(...submittingClasses);
            submitBtn.classList.remove(...normalClasses);
        } else {
            hide(submitBtnSpinner);
            show(submitBtnText);
            submitBtn.removeAttribute("disabled");
            submitBtn.classList.remove(...submittingClasses);
            submitBtn.classList.add(...normalClasses);
        }
    }

    /**
     * Fetch ascii output from backend
     * 
     * @param {HTMLFormElement} form Form element with user selections.
     */
    async function getOutput(form) {
        const action = form.action;
        const method = form.method;
        const formData = new FormData(form);
        formData.delete('size');

        let response = await fetch(action, {
            method,
            body: formData
        });
        let data = await response.json();

        output.textContent = '';
        if (response.status !== 200) {
            throw new Error(data.error);
        }

        removeErrorMessage();

        output.replaceChildren();
        data.forEach(asciiRow => {
            const row = document.createElement("tr");
            for (const c of asciiRow) {
                const cell = document.createElement("td");
                cell.textContent = c;
                row.appendChild(cell);
            }
            output.appendChild(row);
        });

        show(outputContainer);
        outputContainer.tabIndex = "0";
        outputContainer.classList.add(FLOAT_IN_ANIMATION);
    }

    /* ===== ACTIONS ===== */

    /**
     * Toggles between dark mode and light mode
     */
    function displayBtnClickAction() {
        const html = document.documentElement; 
        if (html.classList.contains("dark")) {
            html.classList.remove("dark");
            show(displayLight);
            hide(displayDark);
            show(logo);
            hide(darkLogo);
            displayBtn.title = "Dark theme";
            localStorage.setItem(THEME, LIGHT_THEME);
            themeInput.value = LIGHT_THEME;
        } else {
            html.classList.add("dark");
            show(displayDark);
            hide(displayLight);
            show(darkLogo);
            hide(logo);
            displayBtn.title = "Light theme";
            localStorage.setItem(THEME, DARK_THEME);
            themeInput.value = DARK_THEME;
        }
    }

    function infoBtnClickAction() {
        infoPopup.showModal();
        infoPopup.classList.add(FLOAT_IN_ANIMATION);
    }

    function infoCloseBtnClickAction() {
        infoPopup.close();
        infoPopup.classList.remove(FLOAT_IN_ANIMATION);
    }

    /**
     * Uploads files when dropped into upload button.
     * 
     * @param {DragEvent} event Triggers on drop.
     */
    function uploadBtnDropAction(event) {
        event.preventDefault();
        
        if (event.dataTransfer.files.length === 1) {
            imageInput.files = event.dataTransfer.files;
            const changeEvent = new Event('change');
            imageInput.dispatchEvent(changeEvent);
        } else {
            hideOptions('You can only upload one image at a time.');
        }
    }

    /**
     * Handle new image on upload.
     * 
     * @param {Event} event Triggers on change.
     */
    function imageInputChangeAction(event) {
        const input = event.target;
        const files = input.files;

        if (files.length === 0) {
            hideOptions('No image selected.');
            return;
        }

        handleNewImage(files);
    }

    /**
     * Handles when user clicks size radio.
     * 
     * @param {MouseEvent} event Triggers on click. 
     */
    function sizeRadioClickAction(event) {
        const changeUsability = enabling => {
            let disabledClasses = ['bg-gray-100', 'dark:bg-gray-700'];
            let enabledClasses = ['dark:bg-neutral-900'];
            
            for (const input of widthAndHeightInputs) {
                if (enabling) {
                    input.removeAttribute('readonly');
                    input.setAttribute('tabindex', '0');
                    input.classList.remove(...disabledClasses);
                    input.classList.add(...enabledClasses);
                } else {
                    input.setAttribute('readonly', 'readonly');
                    input.setAttribute('tabindex', '-1');
                    input.classList.remove(...enabledClasses);
                    input.classList.add(...disabledClasses);
                }
            }

            disabledClasses = ['checked:bg-gray-500', 'opacity-40', 'dark:checked:bg-gray-400', 'dark:opacity-70'];
            enabledClasses = ['checked:bg-black', 'dark:checked:bg-white'];

            if (enabling) {
                maintainAspectRatio.classList.remove(...disabledClasses);
                maintainAspectRatio.classList.add(...enabledClasses)
                maintainAspectRatio.removeAttribute('disabled');
            } else {
                maintainAspectRatio.classList.remove(...enabledClasses);
                maintainAspectRatio.classList.add(...disabledClasses)
                maintainAspectRatio.setAttribute('disabled', 'disabled');
                maintainAspectRatio.checked = true;
            }
        }

        const type = event.target.value;
        let width, height, maxHeight; 
        if (type === "custom") {
            changeUsability(true);
            width = parseInt(widthInput.value);
            height = parseInt(heightInput.value);
            maxHeight = MAX_LENGTH;
        } else {
            changeUsability(false);
            width = size[type].width;
            height = size[type].height;
            maxHeight = size[type].maxHeight;
        }
        updateWidthAndHeight(width, height, maxHeight);
    }

    /**
     * Handles when user wants to select a size radio using keyboard (spacebar).
     * 
     * @param {KeyboardEvent} event Triggers on keydown when selecting a size radio. 
     */
    function sizeRadioLabelKeydownAction(event) {
        if (event.key === " ") {
            event.preventDefault();
            event.target.click();
            event.target.focus();
        }
    }

    /**
     * Handles when user submits form.
     * 
     * @param {SubmitEvent} event Triggers on form submit.
     */
    async function formSubmitAction(event) {
        event.preventDefault();
        const form = event.target;

        setSubmitting(true);
        try {
            await getOutput(form);
        } catch(error) {
            console.log(error);
            addErrorMessage(error.message);
        } finally {
            setSubmitting(false);
        }
    }

    /**
     * Grab all the text from the output table, and place into a string.
     * 
     * @returns {string}
     */
    function getOutputText() {
        const rows = output.getElementsByTagName('tr');

        const textRows = Array.from(rows).map(row => {
            const cells = row.getElementsByTagName('td');
            return Array.from(cells).map(cell => cell.textContent).join("");
        });

        return textRows.join("\n");
    }

    /**
     * Handles when user clicks output to add to clipboard.
     */
    async function outputClickAction() {
        const popIn = element => {
            show(element);
            element.classList.add('animate-popin');
        }

        const popOut = element => {
            element.classList.remove('animate-popin');
            element.classList.add('animate-popout');
        }

        clearTimeout(clipboardModalTimeout);

        const text = getOutputText();
        const type = "text/plain";
        const blob = new Blob([text], {type});
        const data = [new ClipboardItem({[type]: blob})];

        try {
            await navigator.clipboard.write(data);
            popIn(copySuccess);
            clipboardModalTimeout = setTimeout(() => popOut(copySuccess), 1500);
        } catch (error) {
            popIn(copyError);
            clipboardModalTimeout = setTimeout(() => popOut(copyError), 1500);
        }
    }

    /**
     * Handles when user scrolls output container
     * 
     * @param {Event} event 
     */
    function outputContainerScrollAction(event) {
        const translation = `translateX(${event.target.scrollLeft}px)`;
        copySuccess.style.transform = translation;
        copyError.style.transform = translation;
    }

    /**
     * Handles when popout animation ends on output overlay.
     * 
     * @param {AnimationEvent} event Triggers on animationend.
     */
    function outputOverlayAnimationEndAction(event) {
        const element = event.target;

        if (event.animationName === 'popout') {
            hide(element);
            element.classList.remove('animate-popout');
        }
    }

    /**
     * Handles when user clicks the "maintain aspect ratio" checkbox ON.
     * 
     * @param {MouseEvent} event 
     */
    function maintainAspectRatioClickAction(event) {
        if (event.target.checked) {
            const width = parseInt(widthInput.value);
            const calculatedHeight = getCalculatedHeight(width);
            updateWidthAndHeight(width, calculatedHeight, MAX_LENGTH);
        }
    }

    /**
     * Handles when user makes change to width input.
     * 
     * @param {Event} event 
     */
    function widthInputChangeAction(event) {
        const width = Math.max(1, Math.min(event.target.value, MAX_LENGTH));
        widthInput.value = width;

        if (isAspectRatioMaintained()) {
            const calculatedHeight = getCalculatedHeight(width);
            heightInput.value = handleSizeWarning(calculatedHeight, MAX_LENGTH);
        }
    }

    /**
     * Handles when user makes change to height input.
     * 
     * @param {Event} event 
     */
    function heightInputChangeAction(event) {
        const height = Math.max(1, Math.min(event.target.value, MAX_LENGTH));
        heightInput.value = height;

        if (isAspectRatioMaintained()) {
            const calculatedWidth = getCalculatedWidth(height);
            widthInput.value = handleSizeWarning(calculatedWidth, MAX_LENGTH, false);
        }
    }

    /* ===== EVENT LISTENERS ===== */

    // High-level button events
    displayBtn.addEventListener('click', displayBtnClickAction);
    infoBtn.addEventListener('click', infoBtnClickAction);
    infoCloseBtn.addEventListener('click', infoCloseBtnClickAction);

    // Upload input events
    uploadBtn.addEventListener('keydown', event => event.key === "Enter" ? imageInput.click() : null);
    uploadBtn.addEventListener('drop', uploadBtnDropAction);
    uploadBtn.addEventListener('dragover', event => event.preventDefault());
    imageInput.addEventListener('change', imageInputChangeAction);

    // Size radio events
    sizeRadios.forEach(radio => {
        radio.addEventListener('click', sizeRadioClickAction);
    });
    for (const label of sizeRadioLabels) {
        label.addEventListener('keydown', sizeRadioLabelKeydownAction);
    }

    // Maintain aspect ratio event
    maintainAspectRatio.addEventListener('click', maintainAspectRatioClickAction);

    // Width & height events
    widthInput.addEventListener('change', widthInputChangeAction);
    heightInput.addEventListener('change', heightInputChangeAction);

    // Exposure input events
    exposure.addEventListener('input', event => exposureValue.value = event.target.value);
    exposureValue.addEventListener('change', event => exposure.value = event.target.value);

    // Form events
    form.addEventListener('submit', formSubmitAction);

    // Output events
    output.addEventListener('click', outputClickAction);
    outputContainer.addEventListener('keydown', event => ["Enter", " "].includes(event.key) ? output.click() : null);
    outputContainer.addEventListener('scroll', outputContainerScrollAction);
    copySuccess.addEventListener('animationend', outputOverlayAnimationEndAction);
    copyError.addEventListener('animationend', outputOverlayAnimationEndAction);

    onLoad();
});