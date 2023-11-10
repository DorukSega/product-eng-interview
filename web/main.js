///API SERVER///
const API_SERVER = window.location.protocol !== "file:" ? `${window.location.protocol}//${window.location.hostname}:8080` : "http://127.0.0.1:8080"
////////////////

/** Indiviual Matrix Cell
 *  @typedef {{from_sdk:number, to_sdk:number, count:number, examples: AppInfo[]}} MatrixItem 
 */

/**
 * @typedef {Object} AppInfo
 * @property {number} id 
 * @property {string} name 
 * @property {Object} company_url 
 * @property {string} company_url.String 
 * @property {boolean} company_url.Valid 
 * @property {Object} release_date 
 * @property {string} release_date.Time 
 * @property {boolean} release_date.Valid 
 * @property {number} genre_id 
 * @property {Object} artwork_large_url 
 * @property {string} artwork_large_url.String 
 * @property {boolean} artwork_large_url.Valid 
 * @property {Object} seller_name 
 * @property {string} seller_name.String 
 * @property {boolean} seller_name.Valid 
 * @property {number} five_star_ratings 
 * @property {number} four_star_ratings 
 * @property {number} three_star_ratings 
 * @property {number} two_star_ratings 
 * @property {number} one_star_ratings 
 */

/** 
 *  @typedef {{id:number, name:string, slug:string, url:{String:string, Valid: boolean}, description:{String:string, Valid: boolean}}} Sdk
 */

/** @type {int[]} */
let selected_sdks = [];

/** all potential sdks
 *  @type {Sdk[]} */
let all_sdks = [];

/** @type {MatrixItem[]} */
let matrix_reference = [];

let global_checksum = 1;

window.onload = async function () {
    const all_sdks_json = await (await fetch(API_SERVER + '/get-sdks').catch(err => (alert("Please make sure that API server is running\n" + err)))).json();
    all_sdks = all_sdks_json.body;
    global_checksum = all_sdks_json.checksum;

    const sdk_view = document.getElementById("rside_cont");
    for (const sdk of all_sdks) {
        sdk_view.append(create_sdk_element(sdk.name, sdk.id))
    }

    await draw_matrix();

    document.getElementById("examples").addEventListener("click", function (msev) {
        /** @this HTMLDivElement */
        if (msev.target.id === "examples")
            this.classList.remove("selected");
    })

    setInterval(getChecksum, 10 * 1000);
}

function create_sdk_element(name, id) {
    const div = document.createElement("div")
    div.textContent = name

    div.addEventListener("click", function () {
        /** @this HTMLDivElement */
        this.classList.toggle("selected")
        if (this.classList.contains("selected"))
            selected_sdks.push(id);
        else
            selected_sdks = selected_sdks.filter(x => x != id);
        draw_matrix();
    })
    div.classList.add("sdk")
    return div
}

function get_sdk_by_id(id) {
    return all_sdks.find(x => x.id == id)
}

/** @param {HTMLElement} parent  */
function create_ledger_elements(parent) {
    parent.innerHTML = ""
    for (const sdk of selected_sdks) {
        const div = document.createElement("div")
        div.textContent = get_sdk_by_id(sdk).name
        parent.append(div)
    }
    const div = document.createElement("div")
    div.textContent = "(none)"
    parent.append(div)
}

async function draw_matrix() {
    /** @type {MatrixItem[]} */
    const data_json = await (
        await fetch(API_SERVER + '/post-matrix', {
            method: 'POST',
            body: JSON.stringify(selected_sdks),
        }).catch(err => (alert("Please make sure that API server is running\n" + err)))
    ).json();
    let data = data_json.body;
    global_checksum = data_json.checksum;
    const ledger_left = document.getElementById("l-ledger");
    const ledger_top = document.getElementById("t-ledger");
    create_ledger_elements(ledger_left)
    create_ledger_elements(ledger_top)

    const matrix_el = document.getElementById("matrix");
    if (data.length === undefined)
        data = [data];

    matrix_el.innerHTML = "" // reset
    let xor_sdks = 0;
    if (selected_sdks.length === 1)
        xor_sdks = 0;
    else if (selected_sdks.length > 1)
        xor_sdks = selected_sdks.reduce((acc, current) => acc ^ current, 0);
    for (let row_ind = 0; row_ind < selected_sdks.length + 1; row_ind++) {
        const row_sel = selected_sdks.length != row_ind ? selected_sdks[row_ind] : -xor_sdks;

        let norm_row = data.filter(x => x.from_sdk == row_sel) // get row
        norm_row = normalize_array(norm_row) // normalize

        for (let col_ind = 0; col_ind < selected_sdks.length + 1; col_ind++) {
            const col_sel = selected_sdks.length != col_ind ? selected_sdks[col_ind] : -xor_sdks;
            let cur_sdk = norm_row.find(d => d.from_sdk === row_sel && d.to_sdk === col_sel)
            if (cur_sdk === undefined)
                cur_sdk = { from_sdk: row_sel, to_sdk: col_sel, count: 0 };
            matrix_el.append(create_matrix_cell(cur_sdk));
        }
    }

    let grid_col = "1fr";
    for (let i = 0; i < selected_sdks.length; i++)
        grid_col += " 1fr";
    matrix_el.style.gridTemplateColumns = grid_col;

    matrix_reference.length = 0
    // Todo: don't update data on main diagonal, ie 33 -> 33 outside of negative -> negative
    // since examples won't change on those
    matrix_reference = matrix_reference.concat(...data)

}

/** @param {MatrixItem} cell  */
function create_matrix_cell(cell) {
    const div = document.createElement("div")
    div.textContent = cell.count + "%";

    const adjr = Math.round(255 - (cell.count / 100) * 255);
    const adjg = Math.round(255 - (cell.count / 100) * 180);

    div.style.backgroundColor = `rgb(${adjr}, ${adjg}, 255)`;


    div.addEventListener("click", async function () {
        const exam_element = document.getElementById("examples");
        await show_examples(cell);
        exam_element.classList.add("selected");
    })

    return div
}

/** @param {MatrixItem[]} row  */
function normalize_array(row) {
    let sum = 0
    row.forEach(it => { sum += it.count });

    if (sum === 0) { // none's if every sdk is selected
        return row.map(it => ({
            from_sdk: it.from_sdk,
            to_sdk: it.to_sdk,
            count: 0
        }));
    }

    return row.map(it => ({
        from_sdk: it.from_sdk,
        to_sdk: it.to_sdk,
        count: Math.round((it.count / sum) * 100)
    }));
}

/** @param {MatrixItem} cell  */
async function show_examples(cell) {
    const example_cont = document.getElementById("example_cont");
    example_cont.innerHTML = ""

    /** @type {AppInfo[]} */
    let data = []
    const matrix_cell_ref = matrix_reference.find(x => x.from_sdk === cell.from_sdk && x.to_sdk === cell.to_sdk);
    if (matrix_cell_ref && matrix_cell_ref.examples !== undefined) {
        data = matrix_cell_ref.examples;
    } else {
        const data_json = await (
            await fetch(API_SERVER + '/post-examples', {
                method: 'POST',
                body: JSON.stringify({ from_sdk: cell.from_sdk, to_sdk: cell.to_sdk, sdk_tuple: selected_sdks }),
            }).catch(err => (alert("Please make sure that API server is running\n" + err)))
        ).json();
        data = data_json.body;
        global_checksum = data_json.checksum;

        if (!data) {
            // when there is 0 examples or 0 count
            if (cell.count !== 0) {
                console.log(`${cell.from_sdk} > ${cell.to_sdk}'s examples are null, while count is not`);
            }
            return;
        }

        if (matrix_cell_ref)
            matrix_cell_ref.examples = data;
    }

    if (data.length === 0)
        return;
    for (const app of data) {
        example_cont.insertAdjacentHTML("beforeend", create_example_element(app))
    }
}

/** @param {AppInfo} app  */
function create_example_element(app) {
    const img_url = app.artwork_large_url.Valid === true ? app.artwork_large_url.String : "./noimg_app.png";
    return `<div> <img class="app_artwork" src="${img_url}" alt="${app.name} Artwork" width="100" height="100"> <div class="app_name">${app.name}</div> </div>`;
}

function matrix_reference_get(from_sdk, to_sdk) {
    return matrix_reference.find(x => x.from_sdk === from_sdk && x.to_sdk === to_sdk)
}

async function getChecksum() {
    const checksum_json = await (await fetch(API_SERVER + '/get-checksum').catch(err => (alert("Please make sure that API server is running\n" + err)))).json();
    const checksum = checksum_json.checksum;

    if (checksum === global_checksum)
        return;
    console.log("Checksum changed, reloading data from server")
    await draw_matrix()
    global_checksum = checksum
}