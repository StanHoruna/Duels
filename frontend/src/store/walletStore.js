import {defineStore} from "pinia";
import {ref} from "vue";
import {Connection, PublicKey, Transaction} from "@solana/web3.js";
import {PhantomWalletAdapter} from "@solana/wallet-adapter-phantom";
import {useUserStore} from "./userStore.js";
import {useNotificationStore} from "./notificationStore.js";

const programId = new PublicKey('TokenkegQfeZyiNwAJbNbGKPFXCWuBvf9Ss623VQ5DA')
const associatedTokenProgramId = new PublicKey('ATokenGPvbdGVxr1b2hvZbsiqW5xWH25efTNsLJA8knL')
const mint = new PublicKey('EPjFWdd5AufqSSqeM2qN1xzybapC8G4wEGGkZwyTDt1v');

export const useWalletStore = defineStore("walletStore", () => {
  const userStore = useUserStore();
  const notificationStore = useNotificationStore();

  const connection = ref(null);
  const wallet = ref(null);
  const address = ref('');
  const balanceData = ref({sol: 0, ucdc: 0});

  const getConnection = () => {
    if (!connection.value) {
      connection.value = new Connection(import.meta.env.VITE_RPC_URL);
    }
  }

  const connect = async (isSignIn = true) => {
    const id = String(Date.now() * Math.random());

    try {
      getConnection();

      wallet.value = new PhantomWalletAdapter();

      if (isSignIn) notificationStore.addNotification({
        type: 'loading',
        text: 'Please confirm wallet connection'
      }, 0, id);

      await wallet.value.connect();

      address.value = wallet.value?.publicKey ? wallet.value?.publicKey.toString() : '';

      if (isSignIn) {
        const encodedMessage = new TextEncoder().encode(address.value);
        const signature = await wallet.value.signMessage(encodedMessage);
        const signedMessage = Buffer.from(signature).toString("base64");
        await userStore.signIn(address.value, signedMessage);
      }
    } catch (error) {
      console.log(error, 'error');
      throw error;
    } finally {
      notificationStore.removeNotification(id);
    }
  };

  const getBalance = async () => {
    getConnection();

    const publicKey = wallet.value?.publicKey || new PublicKey(userStore.userData?.public_address);

    const address = getAssociatedTokenAddress(publicKey);

    // const solBalance = await connection.value?.getBalanceAndContext(publicKey);
    // if (solBalance) balanceData.value.sol = +solBalance.value / LAMPORTS_PER_SOL;

    const accountBalance = await connection.value?.getTokenAccountBalance(address);
    if (accountBalance) balanceData.value.ucdc = +accountBalance.value.amount / 10 ** 6;
  };

  const getAssociatedTokenAddress = (owner) => {
    const [address] = PublicKey.findProgramAddressSync(
      [owner.toBuffer(), programId.toBuffer(), mint.toBuffer()],
      associatedTokenProgramId
    );

    return address;
  };

  const sendTx = async (tx) => {
    try {
      await connect(false);

      const rawTx = Buffer.from(tx, 'base64');
      const transaction = Transaction.from(rawTx);

      if (connection.value instanceof Connection) {
        return await wallet.value?.sendTransaction(transaction, connection.value);
      }
    } catch (error) {
      console.error(error);
    }
  };

  const clearStore = async () => {
    connection.value = null;
    if (wallet.value) await wallet.value.disconnect();
    wallet.value = null;
    address.value = null;
    balanceData.value = {sol: 0, ucdc: 0};
  }

  return {
    balanceData,
    connect,
    clearStore,
    getBalance,
    sendTx,
  };
});
