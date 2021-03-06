package server

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/pkg/browser"
	"github.com/steveoc64/wasmgo/cmd/cmdconfig"
	"github.com/steveoc64/wasmgo/cmd/deployer"
)

func Start(cfg *cmdconfig.Config) error {

	var debug io.Writer
	if cfg.Verbose {
		debug = os.Stdout
	} else {
		debug = ioutil.Discard
	}

	dep, err := deployer.New(cfg)
	if err != nil {
		return err
	}

	svr := &server{cfg: cfg, dep: dep, debug: debug}

	s := &http.Server{Addr: fmt.Sprintf(":%d", cfg.Port), Handler: compressor(svr.ServeHTTP)}

	go func() {
		fmt.Fprintf(debug, "Starting server on %s\n", s.Addr)
		if err := s.ListenAndServe(); err != http.ErrServerClosed {
			log.Fatal(err)
		}
	}()

	go func() {
		if cfg.Open {
			browser.OpenURL(fmt.Sprintf("http://localhost:%d/", cfg.Port))
		}
	}()

	// Set up graceful shutdown
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)

	// Wait for shutdown signal
	<-stop

	fmt.Fprintln(debug, "Stopping server")

	return nil
}

type server struct {
	cfg            *cmdconfig.Config
	dep            *deployer.State
	debug          io.Writer
	cached         bool
	contents, hash []byte
}

func (s *server) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	switch {
	case strings.HasSuffix(req.RequestURI, "/favicon.ico"):
		// ignore
	case strings.HasSuffix(req.RequestURI, "/binary.wasm"):
		// binary
		t1 := time.Now()
		w.Header().Set("Content-Type", "application/wasm")
		if !s.cached {
			contents, hash, err := s.dep.Build()
			if err != nil {
				fmt.Fprintf(os.Stderr, "%v\n", err)
			}
			if _, err := io.Copy(w, bytes.NewReader(contents)); err != nil {
				fmt.Fprintf(os.Stderr, "%v\n", err)
			}
			fmt.Fprintf(s.debug, "Compiled WASM binary with hash %x (%s)\n", hash, time.Since(t1))
			if s.cfg.Cache {
				s.cached = true
				s.contents = contents
				s.hash = hash
			}
		} else {
			if _, err := io.Copy(w, bytes.NewReader(s.contents)); err != nil {
				fmt.Fprintf(os.Stderr, "%v\n", err)
			}
			fmt.Fprintf(s.debug, "Compiled WASM binary from cache%x (%s)\n", s.hash, time.Since(t1))
		}
	case strings.HasSuffix(req.RequestURI, "/loader.js"):
		// loader js
		contents, _, err := s.dep.Loader("/binary.wasm")
		if err != nil {
			fmt.Fprintf(os.Stderr, "%v\n", err)
		}
		w.Header().Set("Content-Type", "application/javascript")
		if _, err := io.Copy(w, bytes.NewReader(contents)); err != nil {
			fmt.Fprintf(os.Stderr, "%v\n", err)
		}
	case strings.HasSuffix(req.RequestURI, "/script.js"):
		// script
		w.Header().Set("Content-Type", "application/javascript")
		if _, err := io.Copy(w, bytes.NewBufferString(WasmExec)); err != nil {
			fmt.Fprintf(os.Stderr, "%v\n", err)
		}
	default:
		// index page
		contents, _, err := s.dep.Index("/script.js", "/loader.js", "/binary.wasm")
		if err != nil {
			fmt.Fprintf(os.Stderr, "%v\n", err)
		}
		w.Header().Set("Content-Type", "text/html")
		if _, err := io.Copy(w, bytes.NewReader(contents)); err != nil {
			fmt.Fprintf(os.Stderr, "%v\n", err)
		}
	}
}

const WasmExec = `(()=>{const e="undefined"!=typeof process;if(e){global.require=require,global.fs=require("fs");const e=require("crypto");global.crypto={getRandomValues(t){e.randomFillSync(t)}},global.performance={now(){const[e,t]=process.hrtime();return 1e3*e+t/1e6}};const t=require("util");global.TextEncoder=t.TextEncoder,global.TextDecoder=t.TextDecoder}else{if("undefined"!=typeof window)window.global=window;else{if("undefined"==typeof self)throw new Error("cannot export Go (neither window nor self is defined)");self.global=self}let e="";global.fs={constants:{O_WRONLY:-1,O_RDWR:-1,O_CREAT:-1,O_TRUNC:-1,O_APPEND:-1,O_EXCL:-1},writeSync(t,n){const i=(e+=s.decode(n)).lastIndexOf("\n");return-1!=i&&(console.log(e.substr(0,i)),e=e.substr(i+1)),n.length},openSync(e,t,s){const n=new Error("not implemented");throw n.code="ENOSYS",n}}}const t=new TextEncoder("utf-8"),s=new TextDecoder("utf-8");if(global.Go=class{constructor(){this.argv=["js"],this.env={},this.exit=(e=>{0!==e&&console.warn("exit code:",e)}),this._callbackTimeouts=new Map,this._nextCallbackTimeoutID=1;const e=()=>new DataView(this._inst.exports.mem.buffer),n=(t,s)=>{e().setUint32(t+0,s,!0),e().setUint32(t+4,Math.floor(s/4294967296),!0)},i=t=>{return e().getUint32(t+0,!0)+4294967296*e().getInt32(t+4,!0)},r=t=>{const s=e().getFloat64(t,!0);if(!isNaN(s))return s;const n=e().getUint32(t,!0);return this._values[n]},o=(t,s)=>{if("number"==typeof s)return isNaN(s)?(e().setUint32(t+4,2146959360,!0),void e().setUint32(t,0,!0)):void e().setFloat64(t,s,!0);switch(s){case void 0:return e().setUint32(t+4,2146959360,!0),void e().setUint32(t,1,!0);case null:return e().setUint32(t+4,2146959360,!0),void e().setUint32(t,2,!0);case!0:return e().setUint32(t+4,2146959360,!0),void e().setUint32(t,3,!0);case!1:return e().setUint32(t+4,2146959360,!0),void e().setUint32(t,4,!0)}let n=this._refs.get(s);void 0===n&&(n=this._values.length,this._values.push(s),this._refs.set(s,n));let i=0;switch(typeof s){case"string":i=1;break;case"symbol":i=2;break;case"function":i=3}e().setUint32(t+4,2146959360|i,!0),e().setUint32(t,n,!0)},a=e=>{const t=i(e+0),s=i(e+8);return new Uint8Array(this._inst.exports.mem.buffer,t,s)},l=e=>{const t=i(e+0),s=i(e+8),n=new Array(s);for(let e=0;e<s;e++)n[e]=r(t+8*e);return n},c=e=>{const t=i(e+0),n=i(e+8);return s.decode(new DataView(this._inst.exports.mem.buffer,t,n))},u=Date.now()-performance.now();this.importObject={go:{"runtime.wasmExit":t=>{const s=e().getInt32(t+8,!0);this.exited=!0,delete this._inst,delete this._values,delete this._refs,this.exit(s)},"runtime.wasmWrite":t=>{const s=i(t+8),n=i(t+16),r=e().getInt32(t+24,!0);fs.writeSync(s,new Uint8Array(this._inst.exports.mem.buffer,n,r))},"runtime.nanotime":e=>{n(e+8,1e6*(u+performance.now()))},"runtime.walltime":t=>{const s=(new Date).getTime();n(t+8,s/1e3),e().setInt32(t+16,s%1e3*1e6,!0)},"runtime.scheduleCallback":t=>{const s=this._nextCallbackTimeoutID;this._nextCallbackTimeoutID++,this._callbackTimeouts.set(s,setTimeout(()=>{this._resolveCallbackPromise()},i(t+8)+1)),e().setInt32(t+16,s,!0)},"runtime.clearScheduledCallback":t=>{const s=e().getInt32(t+8,!0);clearTimeout(this._callbackTimeouts.get(s)),this._callbackTimeouts.delete(s)},"runtime.getRandomData":e=>{crypto.getRandomValues(a(e+8))},"syscall/js.stringVal":e=>{o(e+24,c(e+8))},"syscall/js.valueGet":e=>{o(e+32,Reflect.get(r(e+8),c(e+16)))},"syscall/js.valueSet":e=>{Reflect.set(r(e+8),c(e+16),r(e+32))},"syscall/js.valueIndex":e=>{o(e+24,Reflect.get(r(e+8),i(e+16)))},"syscall/js.valueSetIndex":e=>{Reflect.set(r(e+8),i(e+16),r(e+24))},"syscall/js.valueCall":t=>{try{const s=r(t+8),n=Reflect.get(s,c(t+16)),i=l(t+32);o(t+56,Reflect.apply(n,s,i)),e().setUint8(t+64,1)}catch(s){o(t+56,s),e().setUint8(t+64,0)}},"syscall/js.valueInvoke":t=>{try{const s=r(t+8),n=l(t+16);o(t+40,Reflect.apply(s,void 0,n)),e().setUint8(t+48,1)}catch(s){o(t+40,s),e().setUint8(t+48,0)}},"syscall/js.valueNew":t=>{try{const s=r(t+8),n=l(t+16);o(t+40,Reflect.construct(s,n)),e().setUint8(t+48,1)}catch(s){o(t+40,s),e().setUint8(t+48,0)}},"syscall/js.valueLength":e=>{n(e+16,parseInt(r(e+8).length))},"syscall/js.valuePrepareString":e=>{const s=t.encode(String(r(e+8)));o(e+16,s),n(e+24,s.length)},"syscall/js.valueLoadString":e=>{const t=r(e+8);a(e+16).set(t)},"syscall/js.valueInstanceOf":t=>{e().setUint8(t+24,r(t+8)instanceof r(t+16))},debug:e=>{console.log(e)}}}}async run(e){this._inst=e,this._values=[NaN,void 0,null,!0,!1,global,this._inst.exports.mem,this],this._refs=new Map,this._callbackShutdown=!1,this.exited=!1;const s=new DataView(this._inst.exports.mem.buffer);let n=4096;const i=e=>{let i=n;return new Uint8Array(s.buffer,n,e.length+1).set(t.encode(e+"\0")),n+=e.length+(8-e.length%8),i},r=this.argv.length,o=[];this.argv.forEach(e=>{o.push(i(e))});const a=Object.keys(this.env).sort();o.push(a.length),a.forEach(e=>{o.push(i(` + "`" + `${e}=${this.env[e]}` + "`" + `))});const l=n;for(o.forEach(e=>{s.setUint32(n,e,!0),s.setUint32(n+4,0,!0),n+=8});;){const e=new Promise(e=>{this._resolveCallbackPromise=(()=>{if(this.exited)throw new Error("bad callback: Go program has already exited");setTimeout(e,0)})});if(this._inst.exports.run(r,l),this.exited)break;await e}}static _makeCallbackHelper(e,t,s){return function(){t.push({id:e,args:arguments}),s._resolveCallbackPromise()}}static _makeEventCallbackHelper(e,t,s,n){return function(i){e&&i.preventDefault(),t&&i.stopPropagation(),s&&i.stopImmediatePropagation(),n(i)}}},e){process.argv.length<3&&(process.stderr.write("usage: go_js_wasm_exec [wasm binary] [arguments]\n"),process.exit(1));const e=new Go;e.argv=process.argv.slice(2),e.env=process.env,e.exit=process.exit,WebAssembly.instantiate(fs.readFileSync(process.argv[2]),e.importObject).then(t=>(process.on("exit",t=>{0!==t||e.exited||(e._callbackShutdown=!0,e._inst.exports.run())}),e.run(t.instance))).catch(e=>{throw e})}})();`
