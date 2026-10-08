import { createServer } from "node:http"

// Real served Stimulus/Turbo/controller bytes; only Cable delivery is isolated.
// A tiny HTTP fixture avoids the full application's CSS and navigation machinery.
export async function checkSidebarRefresh(browser, base) {
  const manifest = await (await fetch(`${base}/assets/.manifest.json`)).json()
  const sources = new Map()
  for (const [route, asset] of [["/turbo.js", "turbo.js"], ["/stimulus.js", "stimulus.js"], ["/controller.js", "controllers/rooms_list_controller.js"], ["/helpers.js", "helpers/dom_helpers.js"]]) {
    const response = await fetch(`${base}/assets/${manifest[asset].digested_path}`)
    if (!response.ok) throw new Error(`Cannot read served ${asset}`)
    sources.set(route, await response.text())
  }
  sources.set("/cable.js", `export const cable = { async subscribeTo(_identifier, callbacks) {
    window.sidebarCallbacks = callbacks
    return { unsubscribe() { window.unsubscribed = true } }
  } }`)
  const sidebar = '<turbo-frame id="user_sidebar"><div id="direct_rooms_control"><a href="/sidebar" data-turbo-frame="user_sidebar">Cancel changes</a></div></turbo-frame>'
  let pending, requests = 0, cancelled = 0
  const server = createServer((request, response) => {
    const url = new URL(request.url, "http://localhost")
    if (sources.has(url.pathname)) {
      response.writeHead(200, { "Content-Type": "text/javascript" })
      response.end(sources.get(url.pathname))
    } else if (url.pathname === "/sidebar") {
      requests++
      response.writeHead(200, { "Content-Type": "text/html" })
      if (requests === 1) {
        pending = response
        response.write(sidebar)
        response.once("close", () => { if (!response.writableEnded) cancelled++ })
      } else {
        response.end(sidebar)
      }
    } else {
      response.writeHead(200, { "Content-Type": "text/html" })
      response.end(`<!doctype html><html><body>
        <script type="importmap">{"imports":{"@hotwired/stimulus":"/stimulus.js","@hotwired/turbo-rails":"/cable.js","helpers/dom_helpers":"/helpers.js"}}</script>
        <script type="module">
          import "/turbo.js"
          import { Application } from "/stimulus.js"
          import RoomsList from "/controller.js"
          window.Current = { room: { id: 1 } }
          document.addEventListener("turbo:before-fetch-response", () => { window.slowHeaders = true })
          Application.start().register("rooms-list", RoomsList)
        </script>
        <turbo-frame id="user_sidebar" data-controller="rooms-list"></turbo-frame>
      </body></html>`)
    }
  })
  const context = await browser.newContext()
  try {
    await new Promise(resolve => server.listen(0, "127.0.0.1", resolve))
    const page = await context.newPage()
    const errors = []
    page.on("pageerror", error => errors.push(String(error)))
    page.on("console", message => { if (message.type() === "error") errors.push(message.text()) })
    await page.goto(`http://127.0.0.1:${server.address().port}`)
    await page.waitForFunction(() => window.sidebarCallbacks)
    await page.evaluate(() => { document.getElementById("user_sidebar").src = "/sidebar" })
    await page.waitForFunction(() => window.slowHeaders)
    const finishedEarly = await page.evaluate(async () => {
      window.connectionFinished = false
      Promise.resolve(window.sidebarCallbacks.connected()).then(() => { window.connectionFinished = true })
      await Promise.resolve()
      return window.connectionFinished
    })
    if (finishedEarly) throw new Error("Cable reconnect did not wait for the unfinished frame")
    if (requests !== 1 || !pending) throw new Error("Reconnect restarted the unfinished frame")
    pending.end()
    await page.waitForFunction(() => window.connectionFinished)
    await page.evaluate(async () => { await document.getElementById("user_sidebar").loaded })
    if (requests !== 2 || cancelled !== 0) throw new Error(`Expected one completed load and one reload: ${requests} requests, ${cancelled} cancellations`)

    await page.evaluate(() => {
      document.getElementById("direct_rooms_control").innerHTML = '<a href="/sidebar" data-turbo-frame="user_sidebar">Cancel changes</a><input data-autocomplete-target="input" value="unfinished query"><select data-autocomplete-target="select" multiple><option value="7" selected>Recipient</option></select>'
    })
    await page.evaluate(async () => {
      window.sidebarCallbacks.disconnected()
      await window.sidebarCallbacks.connected()
      await document.getElementById("user_sidebar").loaded
      const input = document.querySelector('[data-autocomplete-target="input"]')
      const select = document.querySelector('[data-autocomplete-target="select"]')
      if (input?.value !== "unfinished query" || select?.selectedOptions.length !== 1 || select.value !== "7") {
        throw new Error("Background reconnect discarded the editor or selected recipients")
      }
    })
    await page.getByRole("link", { name: "Cancel changes", exact: true }).click()
    await page.waitForFunction(() => !document.querySelector('[data-autocomplete-target="input"]'))
    await page.evaluate(() => { document.getElementById("user_sidebar").remove() })
    await page.waitForFunction(() => window.unsubscribed)
    if (errors.length) throw new Error(errors.join("\n"))
    console.log("PASS: sidebar reconnect waits for completed frames, preserves active ping edits, allows cancellation, and unsubscribes removed frames.")
  } finally {
    await context.close()
    server.closeAllConnections()
    await new Promise(resolve => server.close(resolve))
  }
}
