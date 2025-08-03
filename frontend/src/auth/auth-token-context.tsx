import React, {
    createContext,
    useContext,
    useEffect,
    useState,
    useCallback,
    ReactNode,
} from "react";
import { api } from "@/lib/axios.ts";

interface Fakultas {
    id: number;
    kode_fakultas: string;
    nama_fakultas: string;
}


interface Prodi {
    id: number;
    kode_prodi: string;
    nama_prodi: string;
    fakultas_id: number;
    fakultas: Fakultas;
    kel_ukt: string;
}

interface ParsedFullData {
    MhswID?: string;
    Nama?: string;
    TahunID?: string;
    angkatan?: string;
    StatusPernikahan?: string;
    [key: string]: any;
}

interface Mahasiswa {
    mhsw_id: string;
    nama: string;
    email: string;
    prodi_id: number;
    full_data: string;
    kel_ukt: string;
    parsed: ParsedFullData;
    prodi: Prodi;
}

interface UserProfile {
    id: number;
    name: string;
    email: string;
    sso_id: string;
    is_active: boolean;
    mahasiswa?: Mahasiswa;
}




interface AuthContextValue {
    token: string | null;
    isLoggedIn: boolean;
    profile: UserProfile | null;
    setProfile: (profile: UserProfile) => void;
    loadProfile: () => Promise<void>;
    login: (token: string) => void;
    logout: () => void;
    confirmLogout: () => void;
    redirectToLogin: () => void;
    redirectToLogout: () => void;
}

const AuthTokenContext = createContext<AuthContextValue | undefined>(undefined);

interface AuthTokenProviderProps {
    tokenKey: string;
    ssoLoginUrl: string;
    ssoLogoutUrl: string;
    children: ReactNode;
}

function parseJwt(token: string): { exp: number } | null {
    try {
        const base64 = token.split(".")[1];
        const payload = JSON.parse(atob(base64));
        return payload;
    } catch {
        return null;
    }
}

function isExpired(token: string): boolean {
    const payload = parseJwt(token);
    if (!payload?.exp) return true;
    const now = Date.now() / 1000;
    return now >= payload.exp;
}

export const AuthTokenProvider = ({
                                      tokenKey,
                                      ssoLoginUrl,
                                      ssoLogoutUrl,
                                      children,
                                  }: AuthTokenProviderProps) => {
    const [token, setToken] = useState<string | null>(() => {
        const saved = localStorage.getItem(tokenKey);
        return saved ?? null;
    });

    const [profile, setProfile] = useState<UserProfile | null>(null);

    const login = useCallback(
        (newToken: string) => {
            localStorage.setItem(tokenKey, newToken);
            setToken(newToken);
        },
        [tokenKey]
    );

    const logout = useCallback(() => {
        localStorage.removeItem(tokenKey);
        setToken(null);
        setProfile(null);
        window.location.href = ssoLogoutUrl;
    }, [tokenKey, ssoLogoutUrl]);

    const confirmLogout = useCallback(() => {
        const confirmed = window.confirm("Apakah Anda yakin ingin logout?");
        if (confirmed) {
            logout();
            window.location.href = ssoLogoutUrl;
        }
    }, [logout, ssoLogoutUrl]);

    const redirectToLogin = useCallback(() => {
        window.location.href = ssoLoginUrl;
    }, [ssoLoginUrl]);

    const redirectToLogout = useCallback(() => {
        window.location.href = ssoLogoutUrl;
    }, [ssoLogoutUrl]);

    const loadProfile = useCallback(async () => {
        if (!token) return;

        try {
            const res = await api.get<UserProfile>(`/v1/me`, {
                headers: {
                    Authorization: `Bearer ${token}`,
                },
            });
            const rawProfile = res.data;

            try {
                if (rawProfile?.mahasiswa?.full_data) {
                    rawProfile.mahasiswa.parsed = JSON.parse(rawProfile.mahasiswa.full_data);
                }
            } catch (err) {
                console.warn("Gagal parse full_data:", err);
                rawProfile.mahasiswa.parsed = {};
            }

            setProfile(rawProfile);

        } catch (err) {
            console.error("Gagal memuat profil:", err);
            if(import.meta.env?.REDIRECT_ON_FAIL_PROFILE == 1) logout(); // optional: logout jika gagal memuat profil
        }
    }, [token, logout]);

    useEffect(() => {
        const interval = setInterval(() => {
            if (token && isExpired(token)) {
                redirectToLogin();
            }
        }, 5000);
        return () => clearInterval(interval);
    }, [token, redirectToLogin]);

    return (
        <AuthTokenContext.Provider
            value={{
                token,
                isLoggedIn: !!token && !isExpired(token),
                profile,
                setProfile,
                loadProfile,
                login,
                logout,
                confirmLogout,
                redirectToLogin,
                redirectToLogout,
            }}
        >
            {children}
        </AuthTokenContext.Provider>
    );
};

export function useAuthToken(): AuthContextValue {
    const ctx = useContext(AuthTokenContext);
    if (!ctx) throw new Error("useAuthToken must be used within AuthTokenProvider");
    return ctx;
}
