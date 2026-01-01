import { useEffect } from "react";
import { useNavigate } from "react-router-dom";
import { useAuthToken } from "@/auth/auth-token-context";

export default function AuthCallback() {
    const { login, redirectToLogin } = useAuthToken();
    const navigate = useNavigate();

    useEffect(() => {
        const params = new URLSearchParams(window.location.search);
        const token = params.get("token");

        if (token) {
            login(token);
            navigate("/"); // ganti ke halaman tujuan Anda
        } else {
            redirectToLogin(); // jika tidak ada token, redirect ulang
        }
    }, [login, redirectToLogin, navigate]);

    return <p>Logging in...</p>;
}
