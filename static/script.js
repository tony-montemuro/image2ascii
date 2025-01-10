document.addEventListener('DOMContentLoaded', function() {
    /* ===== ELEMENTS ===== */
    const imageInput = document.getElementById('image');
    const imageOptions = document.getElementById('options');
    const customSize = document.getElementById('custom-size');
    const toggleCustomSize = document.getElementById('toggle-custom-size');
    const twitchBtn = document.getElementById('twitch');
    const discordBtn = document.getElementById('discord');
    const widthInput = document.getElementById('width');
    const heightInput = document.getElementById('height');

    /* ===== VARIABLES ===== */
    const size = {
        twitch: {
            width: 30,
            height: undefined
        },
        discord: {
            width: 33,
            height: undefined
        }
    };

    /* ===== FUNCTIONS ===== */
    function checkImage(image) {
        const imageWidth = image.width, imageHeight = image.height;
        size.twitch.height = Math.round((size.twitch.width * imageHeight) / imageWidth / 2);
        size.discord.height = Math.round((size.discord.width * imageHeight) / imageWidth / 2);

        imageOptions.classList.remove('opacity-0');
    };

    function handleNewImage(input, files) {
        const img = files[0];
        const validTypes = ['image/jpeg', 'image/png'];

        if (validTypes.includes(img.type)) {
            const image = new Image();
            image.src = URL.createObjectURL(img);
            image.onload = function() {
                checkImage(this);
            }
        } else {
            imageOptions.classList.add('opacity-0');
            input.value = '';
        }
    };

    function updateSizeInputs(type) {
        widthInput.value = size[type].width;
        heightInput.value = size[type].height;
    };

    /* ===== EVENT LISTENERS ===== */
    imageInput.addEventListener('change', event => {
        const input = event.target;
        const files = input.files;

        if (files.length > 0) {
            handleNewImage(input, files);
        }
    });

    toggleCustomSize.addEventListener('click', event => {
        const checked = event.target.checked;
        const hiddenClasses = ['sr-only', 'w-0', 'h-0', 'absolute'];

        if (checked) {
            hiddenClasses.forEach(hiddenClass => {
                customSize.classList.remove(hiddenClass);
            });
        } else {
            hiddenClasses.forEach(hiddenClass => {
                customSize.classList.add(hiddenClass);
            });
        }
    });

    twitchBtn.addEventListener('click', () => updateSizeInputs('twitch'));
    discordBtn.addEventListener('click', () => updateSizeInputs('discord'));
});