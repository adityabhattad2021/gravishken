import * as types from "@common/types.ts"
import { Mutex } from "./utils";

export function exhausted(d: never) {
    console.log(d)
    throw new Error("unreachable: " + JSON.stringify(d));
}

export type Message = {
    Typ: types.Varient.ExeNotFound,
    Val: types.TExeNotFound,
} | {
    Typ: types.Varient.UserLogin,
    Val: types.TUserLogin,
} | {
    Typ: types.Varient.Quit,
    Val: types.TQuit,
} | {
    Typ: types.Varient.LoadRoute,
    Val: types.TLoadRoute,
} | {
    Typ: types.Varient.WarnUser,
    Val: types.TWarnUser,
} | {
    Typ: types.Varient.ReloadUi,
    Val: types.TReloadUi,
} | {
    Typ: types.Varient.Err,
    Val: types.TErr,
} | {
    Typ: types.Varient.GetTest,
    Val: types.TGetTest,
} | {
    Typ: types.Varient.OpenApp,
    Val: types.TOpenApp,
} | {
    Typ: types.Varient.QuitApp,
    Val: types.TQuitApp,
} | {
    Typ: types.Varient.Unknown,
    Val: unknown,
}

type ValType<T extends types.Varient> = Message extends infer P ? P extends { Typ: T, Val: infer V } ? V : never : never;
type Callback<T extends types.Varient> = (res: ValType<T>) => PromiseLike<void>;
type DisableCallback = () => Promise<void>;

export class Server {
    ws: WebSocket;

    protected wait: Promise<void>;
    protected constructor() {
        this.ws = new WebSocket(`wss://solid-succotash-gwjp9pr7r59265g-6200.app.github.dev/ws`);
        // @ts-ignore
        // this.ws = new WebSocket(`ws://localhost:${import.meta.env.APP_PORT}/ws`);

        this.ws.addEventListener('message', async (e) => {
            let mesg: Message = JSON.parse(e.data);
            // @ts-ignore
            mesg.Val = JSON.parse(mesg.Val);

            await this.handle_message(mesg);
        });

        let resolve: () => {};
        this.wait = new Promise(r => {
            resolve = r as () => {};
        });
        this.ws.addEventListener('open', async (_e) => {
            resolve();
        });
    }

    static async new() {
        let self = new Server();
        await self.wait;
        return self;
    }

    callbacks = new Map<types.Varient, [number, Callback<types.Varient>][]>();
    callback_mutex: Mutex = new Mutex();
    callback_id = 1;
    async add_callback<T extends types.Varient>(type: T, cb: Callback<typeof type>): Promise<DisableCallback> {
        let id = await this.callback_mutex.runExclusive(async () => {
            let id = this.callback_id++;

            // @ts-ignore
            let callback: Callback<types.Varient> = cb;
            let callbacks = this.callbacks.get(type) ?? null;
            if (callbacks == null) {
                this.callbacks.set(type, [[id, callback]]);
            } else {
                callbacks.push([id, callback]);
            }

            return id;
        });

        return async () => {
            await this.callback_mutex.runExclusive(async () => {
                let callbacks = this.callbacks.get(type) ?? [];
                callbacks = callbacks.filter(e => e[0] != id);
                this.callbacks.set(type, callbacks);
            });
        };
    }
    async handle_message(msg: Message) {
        console.log(msg);

        let callbacks = this.callbacks.get(msg.Typ) ?? [];
        for (let callback of callbacks) {
            await callback[1](msg.Val);
        }
        if (callbacks.length > 0) {
            return;
        }

        switch (msg.Typ) {
            case types.Varient.ExeNotFound:
            case types.Varient.LoadRoute:
            case types.Varient.Err:
            case types.Varient.WarnUser:
            case types.Varient.GetTest:
                break;
            case types.Varient.ReloadUi:
                window.location.href = "/";
                break;
            case types.Varient.Quit:
            case types.Varient.OpenApp:
            case types.Varient.QuitApp:
                // redirect this message back to the app :|
                this.send_message(msg);
                break;
            case types.Varient.Unknown:
            case types.Varient.UserLogin: {
                throw new Error(`message type '${msg.Typ}' can't be handled here`);
            } break;
            default:
                throw exhausted(msg);
        }
    }

    send_message(msg: Message) {
        console.log(msg);
        msg.Val = JSON.stringify(msg.Val);
        let json = JSON.stringify(msg);
        this.ws.send(json);
    }
}

// @ts-ignore
export let server: Server = null;
export async function init() {
    server = await Server.new();
}
