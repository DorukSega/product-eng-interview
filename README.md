# Product Eng Interview

## Run Instructions

1. Clone the project.

2. Start the Nix shell:

```sh
nix-shell
```

3. The API will be built and run within the Nix shell environment.

The web/frontend is located in the `./web` directory and does not have any build steps.

4. Open the `index.html` file in a web browser.

## Build The API
Running the following within the same nix-shell environment will build the API as an executable called wapi.

```sh
cd ./api 
go build
 ```
