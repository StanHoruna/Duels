import {apiInstance} from "./instance.js";

// auth
export const SignIn = async (address, secret) => apiInstance.post(`/auth/sign-in-wallet`,{ address, secret });

// user
export const GetUser = async () => apiInstance.get(`/user`);
export const GetUserStats = async () => apiInstance.get(`/user/stats`);
export const UploadAvatar = async (data) => apiInstance.put(`/user/profile-picture`, data);
export const UpdateUsername = async (username) => apiInstance.put(`/user/username`, { username });
export const UploadFile = async (data) => apiInstance.put(`/user/upload-images`, data);

// duel
export const SignCreateDuel = async (duel) => apiInstance.post(`/crypto-duel/solana/sign-tx`, duel);
export const CreateDuel = async (duel, tx_hash) => apiInstance.post(`/crypto-duel/solana`, { ...duel, tx_hash });

export const SignJoinDuel = async (duel_id, answer) => apiInstance.post(`/crypto-duel/solana/join/sign-tx`, { duel_id, answer });
export const JoinDuel = async (duel_id, answer, tx_hash) => apiInstance.post(`/crypto-duel/solana/join`, { duel_id, answer, tx_hash });

export const ResolveDuel = async (duel_id, answer) => apiInstance.put(`/crypto-duel/solana/resolve`, { duel_id, answer });

export const GetDuels = async () => apiInstance.get(`/duel/all`, {
  params: {
    opts: {
      order: {
        order_by: 'created_at',
        order_type: 'desc',
      }
    }
  }
});
export const GetDuelsPublic = async () => apiInstance.get(`/duel/public/all`, {
  params: {
    opts: {
      order: {
        order_by: 'created_at',
        order_type: 'desc',
      }
    }
  }
});

export const GetMyDuels = async () => apiInstance.get(`/duel/my`);
export const GetMyDuelsAsParticipant = async () => apiInstance.get(`/duel/my/participant`, {
  params: {
    opts: {
      order: {
        order_by: 'created_at',
        order_type: 'desc',
      }
    }
  }
});
export const GetDuelByID = async (id) => apiInstance.get(`/duel/${id}`);
export const GetDuelByIDPublic = async (id) => apiInstance.get(`/duel/public/${id}`);
