import {useLocation, useSearchParams} from "react-router-dom";
import { useEffect } from "react";

const NotFound = () => {
  const location = useLocation();
  const [ searchParams ] = useSearchParams();

  const code = searchParams.get("code") ?? "404";
  const error = searchParams.get("error") ?? "Oops! Page not found";
  const errorUrl = import.meta.env.VITE_EPNBP_URL ?? "/";

  useEffect(() => {
    console.error(
      "404 Error: User attempted to access non-existent route:",
      location.pathname
    );
  }, [location.pathname]);
  return (
    <div className="min-h-screen flex items-center justify-center bg-gray-100">
      <div className="text-center">
        <h1 className="text-4xl font-bold mb-4">{code}</h1>
        <p className="text-xl text-gray-600 mb-4">{error}</p>
        <a href={errorUrl} className="text-blue-500 hover:text-blue-700 underline">
          Return to Dashboard
        </a>
      </div>
    </div>
  );
};

export default NotFound;
