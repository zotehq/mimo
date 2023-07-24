# krofi

Krofi is a straightforward image proxy that efficiently handles image resizing and transformations with ease.


## HTTP API Routes

- **GET `/health`**: Retrieve health statistics.
- **GET `/image/resize`**: Resize images.
- **GET `/image/webp`**: Serve images in WebP format.

## `/image/resize`

To use image resizing and transformation, make a GET request to the `/image/resize` endpoint with the following query parameters:

- `url` (required): The URL of the image you want to resize and transform.

Optional Parameters for Image Manipulation:

- `width`: The desired width of the output image (in pixels).
- `height`: The desired height of the output image (in pixels).
- `blur`: The intensity of the Gaussian blur to apply to the image.
- `sharpen`: The intensity of the sharpening effect to apply to the image.
- `gamma`: The gamma correction value to adjust image brightness.
- `contrast`: The contrast adjustment value for the image.
- `brightness`: The brightness adjustment value for the image.
- `saturation`: The saturation adjustment value for the image.

The server will respond with the transformed image in JPEG format, along with appropriate HTTP headers.

## `/image/webp`

To obtain a WebP version of an image, make a GET request to the `/image/webp` endpoint with the following query parameter:

- `url` (required): The URL of the image you want to serve in WebP format.

## License

This project is licensed under the [MIT License](./LICENSE), granting users the freedom to utilize and modify the software as needed.