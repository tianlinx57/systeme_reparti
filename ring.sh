# Fonction de nettoyage
nettoyer () {
  # Suppression des processus de l'application app
  killall app 2> /dev/null

  # Suppression des processus de l'application ctl
  killall ctl 2> /dev/null
  killall prog 2> /dev/null

  # Suppression des processus tee et cat
  killall tee 2> /dev/null
  killall cat 2> /dev/null

  # Suppression des tubes nommés
  \rm -f /tmp/in* /tmp/out*
  \rm -f error.log
  exit 0
}

# Appel de la fonction nettoyer à la réception d'un signal



mkfifo /tmp/in_A /tmp/in_B /tmp/in_C
mkfifo /tmp/out_A /tmp/out_B /tmp/out_C
./prog -n A -t 1 < /tmp/in_A > /tmp/out_A &
./prog -n B -t 2 < /tmp/in_B > /tmp/out_B &
./prog -n C -t 3 < /tmp/in_C > /tmp/out_C &
cat /tmp/out_A > /tmp/in_B &
cat /tmp/out_B > /tmp/in_C &
cat /tmp/out_C > /tmp/in_A &

echo "+ INT QUIT TERM => nettoyer"
trap nettoyer INT QUIT TERM
echo "+ Attente de Crtl C pendant 1h..."
sleep 3600
nettoyer