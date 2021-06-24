export interface CustomWindow extends Window {
    services(): Promise<Tunnel[]>

    toggle(name: string, status: boolean): any
}

export type Tunnel = {
    name: string
    running: boolean
}