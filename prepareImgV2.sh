#!/bin/bash
#set -x

REQ_PKGS=("qemu-utils" "libguestfs-tools")
MNT_POINT="/mnt/image"

#check kernel version : 4.15
if ! uname -a | grep 4.15; then
  echo "Kernel version is not supported"
  exit
fi

usage()
{
  echo "Usage: $0 [ -c | --clean ] [ -r | --runz DIRPATH ]
                  [ -k | --kernel FILEPATH.tgz     ]
                  [ -z | --zpdt   FILEPATH.tar.gz   ] BASE_IMG"
  exit 2
}

# unmount dir & remove tmp files
cleanup() {
    trap - EXIT
    echo "Cleanup .."
    umount ${MNT_POINT}/proc
#    umount ${MNT_POINT}/sys
#    umount ${MNT_POINT}/dev/pts
#    umount ${MNT_POINT}/dev
    umount -f -l ${MNT_POINT} >/dev/null 2>&1
    rmdir ${MNT_POINT} >/dev/null 2>&1
    echo "done"
    exit 0
}

args=$(getopt -a -n $0 -o r:k:z:hc --long runz:,kernel:,zpdt:help,clean -- "$@")
if [ "$?" != "0" ]; then
  usage
fi

eval set -- "$args"
while :
do
  case "$1" in
    -r | --runz)    RUNZ_BIN=`readlink -f $2` ; shift 2 ;;
    -k | --kernel)  KERNEL=`readlink -f $2`   ; shift 2 ;;
    -z | --zpdt)    ZPDT=`readlink -f $2`     ; shift 2 ;;
    -h | --help)    usage ;;
    -c | --clean)   cleanup ;;
    --)             shift; break ;;
    *)              echo "Unexpected option: $1 ."
                    usage ;;
  esac
done

BASE_IMG=`readlink -f $1`

if [[ -z $RUNZ_BIN || -z $KERNEL || -z $ZPDT || -z $BASE_IMG ]]; then
    usage
fi

if [[ $EUID -ne 0 ]]; then
   echo "This script must be run as root"
   exit 1
fi

echo "RUNZ_BIN  : $RUNZ_BIN"
echo "KERNEL    : $KERNEL "
echo "ZPDT      : $ZPDT"
echo "BASE_IMG  : $BASE_IMG"

if [[ ! -d ${RUNZ_BIN} || ! -f ${KERNEL} || ! -f ${ZPDT} || ! -f ${BASE_IMG} ]]; then
    echo "Required dir/file(s) are missing"
    exit 1
fi

if [ ! ${KERNEL##*.} = "tgz" ]; then
    echo "KERNEL : $KERNEL is not the expected one"
    exit 1
fi

if [ ! ${ZPDT##*.} = "gz" ]; then
    echo "ZPDT : $ZPDT is not the expected one"
    exit 1
fi

if [[ ! ${BASE_IMG##*.} = "img" && ! ${BASE_IMG##*.} == "qcow2" ]]; then
    echo "BASE_IMG : $BASE_IMG is not the expected one"
    exit 1
fi
for pkg in ${REQ_PKGS[*]}; do
    dpkg -l | grep -q "$pkg"
    if [ $? -ne 0 ]; then
        echo "$pkg package is missing"
        exit 1
    fi
done

virt-customize -a ${BASE_IMG} --root-password password:coolpass
if [ $? -ne 0 ]; then
    echo "virt-customize failed"
    exit 1
fi

qemu-img resize ${BASE_IMG} 5G
if [ $? -ne 0 ]; then
    echo "qemu-img failed"
    exit 1
fi

mkdir -p ${MNT_POINT}

# Convert cloud image to raw
qemu-img convert -O raw ${BASE_IMG} ${BASE_IMG}.raw || exit 1

# get partition details & mount it
start=`fdisk -l ${BASE_IMG}.raw  | tail -1 | awk '{ print $3 }'`
startoffset=`expr $start \* 512`
mount -o loop,offset=${startoffset}  ${BASE_IMG}.raw ${MNT_POINT} || exit 1

# place all the runz binaries
cp -r ${RUNZ_BIN}/* ${MNT_POINT}/ || exit 1

# create directories
mkdir -p ${MNT_POINT}/volumes
mkdir -p ${MNT_POINT}/container

# extract zPDT binaries
rm -rf ${MNT_POINT}/usr/z1090
mkdir -p ${MNT_POINT}/usr/z1090
tar -C ${MNT_POINT}/usr/z1090/ -zxf ${ZPDT} || exit 1
mv ${MNT_POINT}/usr/z1090/zpdtbin-* ${MNT_POINT}/usr/z1090/bin || exit

# add env. variables
entry='LD_LIBRARY_PATH=/usr/z1090/bin'
grep -q "${entry}" ${MNT_POINT}/etc/environment || echo ${entry} >> ${MNT_POINT}/etc/environment

#Setting memory lock to unlimited
entry='runz      -   memlock     unlimited'
grep -q "$entry" ${MNT_POINT}/etc/security/limits.conf || echo "$entry" >> ${MNT_POINT}/etc/security/limits.conf
entry='root      -   memlock     unlimited'
grep -q "$entry" ${MNT_POINT}/etc/security/limits.conf || echo "$entry" >> ${MNT_POINT}/etc/security/limits.conf

# extract kernel patches
mkdir -p ${MNT_POINT}/patches
tar -C ${MNT_POINT}/patches -zxf ${KERNEL} || exit 1
KVER=`basename /mnt/image/patches/kernel-* | awk -F'-' '{ print $2"-"$3}'`
PATCHDIR="/patches/kernel-${KVER}"
KVER=${KVER%.*}

echo "PATCHDIR: ${PATCHDIR}"
echo "KVER: ${KVER}"

trap cleanup EXIT
mount -t proc -o nosuid,noexec,nodev proc ${MNT_POINT}/proc
# mount -t sysfs -o nosuid,noexec,nodev sysfs ${MNT_POINT}/sys
# mount -o bind /dev ${MNT_POINT}/dev/
# mkdir -p -m 755 ${MNT_POINT}/dev/pts
# mount -t devtmpfs -o mode=0755,nosuid devtmpfs ${MNT_POINT}/dev
# mount -t devpts -o gid=5,mode=620 devpts ${MNT_POINT}/dev/pts

# chroot ${MNT_POINT} update-initramfs -u -k all

chroot ${MNT_POINT} /bin/bash <<EOT
    set -x

    #check kernel version : 4.15
    if ! uname -a | grep 4.15; then
    echo "Kernel version is not supported"
    exit 1
    fi

	# Add DNS & install dependencies to the kernel patches
	mv /etc/resolv.conf /etc/resolv.conf.bk
	echo 'nameserver 8.8.8.8' | tee /etc/resolv.conf
	apt update  && apt install -y crda binutils libdw1
	mv /etc/resolv.conf.bk /etc/resolv.conf

	# install the kernel patches
	cd ${PATCHDIR} && dpkg -i *.deb

    # remove packages
    rm -rf ${PATCHDIR}

	# Create a user runz with UID&GID 1001
	useradd -u 1001 -m runz

    # provide sudo privileges
    usermod -aG sudo runz

    # Change dir/file ownership
    chown runz:runz /volumes
    chown runz:runz /dev/net/tun
    chown runz:runz /dev/kvm
    chmod 0666 /dev/kvm
EOT

umount ${MNT_POINT}/proc
# umount ${MNT_POINT}/sys
# umount ${MNT_POINT}/dev/pts
# umount ${MNT_POINT}/dev

sync

qemu-img convert -f raw -O qcow2  ${BASE_IMG}.raw ${BASE_IMG}.qcow2

trap - EXIT
cleanup
