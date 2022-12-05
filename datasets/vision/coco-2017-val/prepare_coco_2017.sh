#!/bin/bash
#Exit immediately if a command exits with a non-zero status.
set -e

python3 -m pip install -r requirements.txt
ck pull repo:ck-env
ck install package --tags=object-detection,dataset,coco,2017,val,original
ck locate env --tags=object-detection,dataset,coco,2017,val,original
python3 preprocess.py --input-val_dir $(ck locate env --tags=object-detection,dataset,coco,2017,val,original | tail -1)/val2017