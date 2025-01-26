document.addEventListener('DOMContentLoaded', function() {
    /* ===== ELEMENTS ===== */
    const form = this.getElementById('form');
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
    const copySuccess = this.getElementById("output-copy-success");
    const copyError = this.getElementById("output-copy-failure");
    const outputContainer = this.getElementById("output-wrapper");
    const submitBtn = this.getElementById("submit");
    const submitBtnText = this.getElementById("submit-btn-text");
    const submitBtnSpinner = this.getElementById("submit-btn-spinner");
    const sizeRadios = sizeContainer.querySelectorAll('input[name="size"]');
    const sizeRadioLabels = sizeContainer.getElementsByTagName('label'); 
    const widthAndHeightInputs = customSize.getElementsByTagName('input');

    /* ===== VARIABLES ===== */
    const size = {
        twitch: {
            width: 30,
            height: undefined,
            maxHeight: 16
        },
        discord: {
            width: 32,
            height: undefined,
            maxHeight: 62
        },
        small: {
            width: 30,
            height: undefined
        },
        medium: {
            width: 90,
            height: undefined
        },
        large: {
            width: 150,
            height: undefined
        }
    };
    let clipboardModalTimeout;

    /* ===== FUNCTIONS ===== */

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

        if (message) {
            addErrorMessage(message);
        }
    }

    /**
     * Update the width and height based on `type`
     * 
     * @param {string} type - Must correspond to a key within 'size'.
     */
    function updateWidthAndHeight(type) {
        widthInput.value = size[type].width;
        heightInput.value = size[type].height;
    };

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

        updateWidthAndHeight(sizeContainer.querySelector('input:checked').value);
    };


    /**
     * Determine height of ascii
     * 
     * @param {number} imageWidth - The width of the original image.
     * @param {number} imageHeight - The height of the original image.
     * @param {string} type - Must correspond to a key within 'size'.
     * @returns {number} The new height.
     */
    function getHeight(imageWidth, imageHeight, type) {
        const maxHeight = size[type].maxHeight ?? Number.MAX_SAFE_INTEGER;
        const calculatedHeight = Math.round((size[type].width * imageHeight) / imageWidth / 2);
        return Math.min(calculatedHeight, maxHeight);
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
            const image = new Image();
            image.src = URL.createObjectURL(img);
            image.onload = function() {
                this.setAttribute('name', img.name);
                Object.keys(size).forEach(type => size[type].height = getHeight(this.width, this.height, type));
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
        data.forEach(row => {
            output.textContent += row + "\n";
        });
        show(outputContainer);
        outputContainer.tabIndex = "0";
    }

    // actions

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
            for (const input of widthAndHeightInputs) {
                if (enabling) {
                    input.removeAttribute('readonly');
                    input.classList.remove('bg-gray-100');
                } else {
                    input.setAttribute('readonly', 'readonly');
                    input.classList.add('bg-gray-100');
                }
            }
        }

        const type = event.target.value;
        if (type === "custom") {
            changeUsability(true);
        } else {
            changeUsability(false);
            updateWidthAndHeight(type);
        }
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
            addErrorMessage(error.message);
        } finally {
            setSubmitting(false);
        }
    }

    /**
     * Handles when user clicks output to add to clipboard.
     * 
     * @param {MouseEvent} event Triggers on click.
     */
    async function outputClickAction(event) {
        const popIn = element => {
            show(element);
            element.classList.add('animate-popin');
        }

        const popOut = element => {
            element.classList.remove('animate-popin');
            element.classList.add('animate-popout');
        }

        clearTimeout(clipboardModalTimeout);

        const element = event.target;
        const text = element.textContent;
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

    /* ===== EVENT LISTENERS ===== */

    // Upload input events
    uploadBtn.addEventListener('keydown', event => event.key === "Enter" ? imageInput.click() : null);
    uploadBtn.addEventListener('drop', event => uploadBtnDropAction(event));
    uploadBtn.addEventListener('dragover', event => event.preventDefault());
    imageInput.addEventListener('change', event => imageInputChangeAction(event));

    // Size radio events
    sizeRadios.forEach(radio => {
        radio.addEventListener('click', event => sizeRadioClickAction(event));
    });
    for (const label of sizeRadioLabels) {
        label.addEventListener('keydown', event => sizeRadioLabelKeydownAction(event));
    }

    // Exposure input events
    exposure.addEventListener('input', event => {
        exposureValue.value = event.target.value;
    });
    exposureValue.addEventListener('change', event => {
        exposure.value = event.target.value;
    });

    // Form events
    form.addEventListener('submit', event => formSubmitAction(event));

    // Output events
    output.addEventListener('click', event => outputClickAction(event));
    outputContainer.addEventListener('keydown', event => ["Enter", " "].includes(event.key) ? output.click() : null);
    copySuccess.addEventListener('animationend', event => outputOverlayAnimationEndAction(event));
    copyError.addEventListener('animationend', event => outputOverlayAnimationEndAction(event));
});