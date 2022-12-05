

## ResNet-50 Datasets
| dataset | download link | 
| ---- | ---- | 
| coco (validation) | http://images.cocodataset.org/zips/val2017.zip | 

### Install dependencies

First install all python dependencies in the following manner:
```bash
$ python3 -m pip install -r requirements.txt
```

### Download using Collective Knowledge (CK)

The recommended way to download the datasets is using the [Collective Knowledge](http://cknowledge.org)
framework (CK) for collaborative and reproducible research.

First, confirm you have ck, and pull its repositories containing dataset packages:
```bash
$ ck version
V1.15.0
$ ck pull repo:ck-env
```

#### Download COCO 2017 validation dataset
```bash
$ ck install package --tags=object-detection,dataset,coco,2017,val,original
$ ck locate env --tags=object-detection,dataset,coco,2017,val,original
```


##### Pre-Process COCO 2017

```bash
$ python3 preprocess.py --input-val_dir $(ck locate env --tags=object-detection,dataset,coco,2017,val,original | tail -1)/val2017
Using random seed 12345 to take a random 224 x 224 crop to the scaled image
Saving cropped scaled images to cropped-val2017
100%|███████████████████████████████████████████████████████████████████████████████████| 5000/5000 [00:08<00:00, 560.21it/s]
```