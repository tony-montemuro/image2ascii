document.addEventListener('DOMContentLoaded', function() {
    const imageInput = document.getElementById('image');
    const imageOptions = document.getElementById('options');

    function handleNewImage(input, files) {
        const image = files[0];
        const validTypes = ['image/jpeg', 'image/png'];

        if (validTypes.includes(image.type)) {
            imageOptions.classList.remove('opacity-0');
        } else {
            imageOptions.classList.add('opacity-0');
            input.value = '';
        }
    };

    imageInput.addEventListener('change', event => {
        const input = event.target;
        const files = input.files;

        if (files.length > 0) {
            handleNewImage(input, files);
        }
    });
});