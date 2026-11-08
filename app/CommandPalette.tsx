"use client";

import { useEffect, useId, useState } from "react";
import { ArrowLeft, GitCompareArrows, Heart, LayoutGrid, Search, ShoppingBag, X, Zap } from "lucide-react";

export type CommandId = "catalog" | "categories" | "cart" | "favorites" | "compare" | "clear" | "language";

export type CommandPaletteLabels = {
  commandPalette: string;
  commandSearch: string;
  noCommands: string;
  commandHint: string;
  commandCatalog: string;
  commandCategories: string;
  commandCart: string;
  commandFavorites: string;
  commandCompare: string;
  commandClear: string;
  commandLanguage: string;
  close: string;
};

type Props = {
  labels: CommandPaletteLabels;
  onClose: () => void;
  onRunCommand: (id: CommandId) => void;
  onReady: () => void;
};

const commandIcons = {
  catalog: Search,
  categories: LayoutGrid,
  cart: ShoppingBag,
  favorites: Heart,
  compare: GitCompareArrows,
  clear: X,
  language: Zap
};

export default function CommandPalette({ labels, onClose, onRunCommand, onReady }: Props) {
  const [query, setQuery] = useState("");
  const [activeIndex, setActiveIndex] = useState(0);
  const inputId = useId();

  useEffect(() => {
    onReady();
  }, [onReady]);

  const commandItems = [
    { id: "catalog", label: labels.commandCatalog, keywords: "shop products storefront catalog ویترین فروشگاه محصول" },
    { id: "categories", label: labels.commandCategories, keywords: "category categories browse دسته بندی گروه" },
    { id: "cart", label: labels.commandCart, keywords: "bag basket cart سبد خرید" },
    { id: "favorites", label: labels.commandFavorites, keywords: "saved favorite wishlist علاقه مندی ذخیره" },
    { id: "compare", label: labels.commandCompare, keywords: "comparison compare مقایسه" },
    { id: "clear", label: labels.commandClear, keywords: "reset filters پاک کردن فیلتر بازنشانی" },
    { id: "language", label: labels.commandLanguage, keywords: "locale فارسی english زبان" }
  ].filter(item => (item.label + " " + item.keywords).toLocaleLowerCase().includes(query.trim().toLocaleLowerCase()));

  const chooseCommand = (id: CommandId) => onRunCommand(id);

  return (
    <div className="overlay modal-overlay command-palette-overlay" onMouseDown={event => {
      if (event.target === event.currentTarget) onClose();
    }}>
      <section className="command-palette" role="dialog" aria-modal="true" aria-labelledby={inputId + "-title"} aria-describedby={inputId + "-hint"} tabIndex={-1}>
        <div className="command-palette-heading">
          <div><span className="section-kicker">GLOBALCART / COMMANDS</span><h2 id={inputId + "-title"}>{labels.commandPalette}</h2></div>
          <button className="icon-button" onClick={onClose} aria-label={labels.close}><X size={18}/></button>
        </div>
        <label className="command-palette-search" htmlFor={inputId}>
          <Search size={18}/>
          <input
            id={inputId}
            autoFocus
            value={query}
            role="combobox"
            aria-autocomplete="list"
            aria-controls={inputId + "-list"}
            aria-expanded={commandItems.length > 0}
            aria-activedescendant={commandItems[activeIndex] ? inputId + "-option-" + activeIndex : undefined}
            onChange={event => { setQuery(event.target.value); setActiveIndex(0); }}
            onKeyDown={event => {
              if (event.key === "ArrowDown") {
                event.preventDefault();
                setActiveIndex(current => commandItems.length ? Math.min(current + 1, commandItems.length - 1) : 0);
              } else if (event.key === "ArrowUp") {
                event.preventDefault();
                setActiveIndex(current => Math.max(current - 1, 0));
              } else if (event.key === "Enter") {
                event.preventDefault();
                const chosen = commandItems[activeIndex] ?? commandItems[0];
                if (chosen) chooseCommand(chosen.id as CommandId);
              }
            }}
            placeholder={labels.commandSearch}
          />
        </label>
        <div className="command-palette-list" id={inputId + "-list"} role="listbox" aria-label={labels.commandPalette}>
          {commandItems.length === 0
            ? <p className="command-palette-empty" role="status">{labels.noCommands}</p>
            : commandItems.map((command, index) => {
                const Icon = commandIcons[command.id as CommandId];
                return (
                  <button
                    type="button"
                    role="option"
                    id={inputId + "-option-" + index}
                    aria-selected={index === activeIndex}
                    className={index === activeIndex ? "command-option active" : "command-option"}
                    key={command.id}
                    onMouseEnter={() => setActiveIndex(index)}
                    onClick={() => chooseCommand(command.id as CommandId)}
                  >
                    <span className="command-option-icon"><Icon size={17}/></span>
                    <b>{command.label}</b>
                    <span className="command-option-key">{index === activeIndex ? "↵" : index + 1}</span>
                  </button>
                );
              })}
        </div>
        <div className="command-palette-footer" id={inputId + "-hint"}>
          <span>{labels.commandHint}</span>
          <span><kbd>↑</kbd><kbd>↓</kbd><kbd>Enter</kbd><kbd>Esc</kbd></span>
        </div>
      </section>
    </div>
  );
}
