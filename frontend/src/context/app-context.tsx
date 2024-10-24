import { useContext, createContext, useState } from "react";


interface StateContextType {
    username: string;
    userPassword: string;
    setUsername: (username: string) => void;
    setUserPassword: (userPassword: string) => void;
}


const StateContext = createContext({} as StateContextType);


interface StateContextProviderProps {
    children: React.ReactNode;
}

export function StateContextProvider({ children }: StateContextProviderProps) {
    const [username, setUsername] = useState('');
    const [userPassword, setUserPassword] = useState('');

    return (
        <StateContext.Provider
            value={{
                username,
                userPassword,
                setUsername,
                setUserPassword,
            }}
        >
            {children}
        </StateContext.Provider>
    )
}

export function useStateContext() {
    return useContext(StateContext);
}
