import { useState } from "react";
import { Button } from "@/components/ui/button";
import { useToast } from "@/components/ui/use-toast";
import axios from "axios";
import { useAuthToken } from "@/auth/auth-token-context";
import { useStudentBills } from "@/bill/context.tsx";

export const SuccessBills = () => {
  return (
      <div className="mb-6">
        <div className="border p-4 rounded text-green-600 border-green-500 bg-green-50">
          <p className="text-sm font-medium">🎉 Anda sudah melunasi semua tagihan untuk semester ini.</p>
        </div>
      </div>
  );
};
