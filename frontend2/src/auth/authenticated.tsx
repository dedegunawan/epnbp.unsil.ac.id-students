import { useAuthToken } from "@/auth/auth-token-context";
import { useEffect } from "react";
import { Outlet } from "react-router-dom";

export default function Authenticated() {
    const { isLoggedIn, redirectToLogin } = useAuthToken();

    useEffect(() => {
        if (!isLoggedIn) {
            const timeout = setTimeout(() => {
                redirectToLogin();
            }, 1000);
            return () => clearTimeout(timeout);
        }
    }, [isLoggedIn, redirectToLogin]);

    return isLoggedIn ? <Outlet /> : <p>Redirecting to login...</p>;
}
