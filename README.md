# 🎮 Power4 Web

Un projet en **Go (Golang)** qui implémente le jeu **Puissance 4** jouable via une interface web locale.  
Deux joueurs peuvent s’affronter en alternant les tours grâce à un serveur HTTP minimaliste et une interface générée avec des **templates HTML**.

---

## 🚀 Fonctionnalités

- Plateau de jeu de **7 colonnes × 6 lignes** affiché en HTML (`<table>` ou `<div>`).
- Gestion de deux joueurs prenant tour à tour la main.
- Possibilité de jouer en sélectionnant une colonne via un bouton ou formulaire HTML.
- Rafraîchissement automatique de la page après chaque coup avec mise à jour du plateau.
- Vérification des conditions de victoire :
  - Alignement de **4 pions horizontaux, verticaux ou diagonaux**.
- Détection de l’égalité si la grille est complètement remplie.

---

## 🛠️ Stack technique

- **Langage :** Go (Golang)  
- **Backend :** Serveur HTTP (`net/http`)  
- **Frontend :** HTML / CSS basique (sans framework)  
- **Templates :** Go `html/template`  

---

## 📌 Endpoints

- **GET /**  
  → Affiche la grille et l’interface du jeu (plateau, infos joueur, état).  

- **POST /play**  
  → Reçoit la colonne choisie par le joueur, met à jour l’état du jeu et recharge l’interface.  

---

## 📂 Structure recommandée du projet